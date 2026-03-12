package main

import (
	"context"
	"encoding/json"
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

	if in.appID == "" && in.appName == "" && !in.fromPRPreview {
		a.Fatalf("either app_id, app_name, or from_pr_preview must be set")
	}

	do := godo.NewFromToken(in.token)
	do.UserAgent = "do-app-action/unarchive"

	appID := in.appID
	if appID == "" {
		appName := in.appName
		if appName == "" {
			ghCtx, err := a.Context()
			if err != nil {
				a.Fatalf("failed to get GitHub context: %v", err)
			}
			repoOwner, repo := ghCtx.Repo()
			prRef, err := utils.PRRefFromContext(ghCtx)
			if err != nil {
				a.Fatalf("failed to get PR number: %v", err)
			}
			appName = utils.GenerateAppName(repoOwner, repo, prRef)
		}

		app, err := utils.FindAppByName(ctx, do.Apps, appName)
		if err != nil {
			a.Fatalf("failed to find app: %v", err)
		}
		if app == nil {
			if in.ignoreNotFound {
				a.Infof("app %q not found, ignoring", appName)
				return
			}
			a.Fatalf("app %q not found", appName)
		}
		appID = app.ID
	}

	// Fetch current app to get the live spec before mutating it.
	app, resp, err := do.Apps.Get(ctx, appID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound && in.ignoreNotFound {
			a.Infof("app %q not found, ignoring", appID)
			return
		}
		a.Fatalf("failed to get app: %v", err)
	}

	// Clear the maintenance/archive flag from the spec.
	spec := app.Spec
	spec.Maintenance = &godo.AppMaintenanceSpec{
		Archive: false,
	}

	updated, _, err := do.Apps.Update(ctx, appID, &godo.AppUpdateRequest{Spec: spec})
	if err != nil {
		a.Fatalf("failed to unarchive app: %v", err)
	}

	a.Infof("App %q unarchived successfully", appID)

	liveURL := ""
	if in.waitForLive {
		liveURL = waitForLiveURL(ctx, a, do, appID, in.waitTimeout)
	}

	appJSON, err := json.Marshal(updated)
	if err != nil {
		a.Warningf("failed to marshal app to JSON: %v", err)
	} else {
		a.SetOutput("app", string(appJSON))
	}
	a.SetOutput("live_url", liveURL)
}

// waitForLiveURL polls until the app has a non-empty LiveURL or timeout is reached.
// A timeoutSecs value of 0 means poll indefinitely until the app is live.
func waitForLiveURL(ctx context.Context, a *gha.Action, do *godo.Client, appID string, timeoutSecs int) string {
	pollInterval := 10 * time.Second

	if timeoutSecs == 0 {
		a.Infof("Waiting indefinitely for app %q to become live...", appID)
	} else {
		a.Infof("Waiting up to %ds for app %q to become live...", timeoutSecs, appID)
	}

	var deadline time.Time
	if timeoutSecs > 0 {
		deadline = time.Now().Add(time.Duration(timeoutSecs) * time.Second)
	}

	for timeoutSecs == 0 || time.Now().Before(deadline) {
		app, _, err := do.Apps.Get(ctx, appID)
		if err != nil {
			a.Warningf("error polling app status: %v", err)
		} else if app.LiveURL != "" {
			a.Infof("App is live at %s", app.LiveURL)
			return app.LiveURL
		}
		time.Sleep(pollInterval)
	}

	a.Warningf("timed out waiting for app %q to become live", appID)
	return ""
}
