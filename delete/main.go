package main

import (
	"context"
	"fmt"
	"net/http"

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
	do.UserAgent = "do-app-action/delete"

	r := &runner{
		action: a,
		apps:   do.Apps,
		inputs: in,
	}
	if err := r.run(ctx); err != nil {
		a.Fatalf("%v", err)
	}
}

type runner struct {
	action *gha.Action
	apps   godo.AppsService
	inputs inputs
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

	if resp, err := r.apps.Delete(ctx, appID); err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound && r.inputs.ignoreNotFound {
			r.action.Infof("app %q not found, ignoring", appID)
			return nil
		}
		return fmt.Errorf("failed to delete app: %w", err)
	}
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
