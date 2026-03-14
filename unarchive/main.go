package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"

	"github.com/blavity/do-app-action/utils"
)

func main() {
	ctx := context.Background()
	a := gha.New()

	in, err := getInputs(a)
	if err != nil {
		a.Fatalf("failed to get inputs: %v", err)
	}
	a.AddMask(in.token)

	do := godo.NewFromToken(in.token)
	do.UserAgent = "do-app-action/unarchive"

	r := &runner{
		action:       a,
		apps:         do.Apps,
		inputs:       in,
		pollInterval: 10 * time.Second,
		now:          time.Now,
		sleep:        time.Sleep,
	}
	if err := r.run(ctx); err != nil {
		a.Fatalf("%v", err)
	}
}

type runner struct {
	action       *gha.Action
	apps         godo.AppsService
	inputs       inputs
	pollInterval time.Duration
	now          func() time.Time
	sleep        func(time.Duration)
}

func (r *runner) run(ctx context.Context) error {
	if r.inputs.appID == "" && r.inputs.appName == "" && !r.inputs.fromPRPreview {
		return fmt.Errorf("either app_id, app_name, or from_pr_preview must be set")
	}

	appID, err := r.resolveAppID(ctx)
	if err != nil {
		return err
	}
	if appID == "" {
		// resolveAppID returned empty — app not found and ignoreNotFound was set.
		return nil
	}

	// Fetch current app to get the live spec before mutating it.
	app, resp, err := r.apps.Get(ctx, appID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound && r.inputs.ignoreNotFound {
			r.action.Infof("app %q not found, ignoring", appID)
			return nil
		}
		return fmt.Errorf("failed to get app: %w", err)
	}

	// Clear the maintenance/archive flag from the spec.
	spec := app.Spec
	spec.Maintenance = &godo.AppMaintenanceSpec{
		Archive: false,
	}

	updated, _, err := r.apps.Update(ctx, appID, &godo.AppUpdateRequest{Spec: spec})
	if err != nil {
		return fmt.Errorf("failed to unarchive app: %w", err)
	}

	r.action.Infof("App %q unarchived successfully", appID)

	liveURL := ""
	if r.inputs.waitForLive {
		liveURL = r.waitForLiveURL(ctx, appID)
	}

	appJSON, err := json.Marshal(updated)
	if err != nil {
		r.action.Warningf("failed to marshal app to JSON: %v", err)
	} else {
		r.action.SetOutput("app", string(appJSON))
	}
	r.action.SetOutput("live_url", liveURL)
	return nil
}

// resolveAppID returns the app ID to operate on. If the app is not found and
// ignoreNotFound is set, it returns ("", nil). If the app is not found and
// ignoreNotFound is not set, it returns an error.
func (r *runner) resolveAppID(ctx context.Context) (string, error) {
	if r.inputs.appID != "" {
		return r.inputs.appID, nil
	}

	appName := r.inputs.appName
	if appName == "" {
		ghCtx, err := r.action.Context()
		if err != nil {
			return "", fmt.Errorf("failed to get GitHub context: %w", err)
		}
		repoOwner, repo := ghCtx.Repo()
		prRef, err := utils.PRRefFromContext(ghCtx)
		if err != nil {
			return "", fmt.Errorf("failed to get PR number: %w", err)
		}
		appName = utils.GenerateAppName(repoOwner, repo, prRef)
	}

	app, err := utils.FindAppByName(ctx, r.apps, appName)
	if err != nil {
		return "", fmt.Errorf("failed to find app: %w", err)
	}
	if app == nil {
		if r.inputs.ignoreNotFound {
			r.action.Infof("app %q not found, ignoring", appName)
			return "", nil
		}
		return "", fmt.Errorf("app %q not found", appName)
	}
	return app.ID, nil
}

// waitForLiveURL polls until the app has a non-empty LiveURL or timeout is reached.
// A waitTimeout value of 0 means poll indefinitely until the app is live.
func (r *runner) waitForLiveURL(ctx context.Context, appID string) string {
	timeoutSecs := r.inputs.waitTimeout
	if timeoutSecs == 0 {
		r.action.Infof("Waiting indefinitely for app %q to become live...", appID)
	} else {
		r.action.Infof("Waiting up to %ds for app %q to become live...", timeoutSecs, appID)
	}

	var deadline time.Time
	if timeoutSecs > 0 {
		deadline = r.now().Add(time.Duration(timeoutSecs) * time.Second)
	}

	for timeoutSecs == 0 || r.now().Before(deadline) {
		app, _, err := r.apps.Get(ctx, appID)
		if err != nil {
			r.action.Warningf("error polling app status: %v", err)
		} else if app.LiveURL != "" {
			r.action.Infof("App is live at %s", app.LiveURL)
			return app.LiveURL
		}
		r.sleep(r.pollInterval)
	}

	r.action.Warningf("timed out waiting for app %q to become live", appID)
	return ""
}
