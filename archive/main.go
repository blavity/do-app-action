package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/blavity/do-app-action/utils"
	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
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
	do.UserAgent = "do-app-action/archive"

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

	// Apply maintenance/archive to the spec.
	spec := app.Spec
	spec.Maintenance = &godo.AppMaintenanceSpec{
		Archive: true,
	}

	updated, _, err := do.Apps.Update(ctx, appID, &godo.AppUpdateRequest{Spec: spec})
	if err != nil {
		a.Fatalf("failed to archive app: %v", err)
	}

	a.Infof("App %q archived successfully", appID)

	appJSON, err := json.Marshal(updated)
	if err != nil {
		a.Warningf("failed to marshal app to JSON: %v", err)
	} else {
		a.SetOutput("app", string(appJSON))
	}
}
