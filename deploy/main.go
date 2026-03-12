package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
	"sigs.k8s.io/yaml"

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
	do.UserAgent = "do-app-action/deploy"
	d := &deployer{
		action:     a,
		apps:       do.Apps,
		httpClient: http.DefaultClient,
		inputs:     in,
	}

	spec, err := d.createSpec(ctx)
	if err != nil {
		a.Fatalf("failed to create spec: %v", err)
	}

	if in.deployPRPreview {
		ghCtx, err := a.Context()
		if err != nil {
			a.Fatalf("failed to get GitHub context: %v", err)
		}
		if err := utils.SanitizeSpecForPullRequestPreview(spec, ghCtx); err != nil {
			a.Fatalf("failed to sanitize spec for PR preview: %v", err)
		}
	}

	app, err := d.deploy(ctx, spec)
	if app != nil {
		appJSON, err := json.Marshal(app)
		if err != nil {
			a.Warningf("failed to marshal app to JSON: %v", err)
		} else {
			a.SetOutput("app", string(appJSON))
		}
	}
	if err != nil {
		a.Fatalf("failed to deploy app: %v", err)
	}
}

type deployer struct {
	action     *gha.Action
	apps       godo.AppsService
	httpClient *http.Client
	inputs     inputs
}

func (d *deployer) createSpec(ctx context.Context) (*godo.AppSpec, error) {
	if d.inputs.appName != "" {
		app, err := utils.FindAppByName(ctx, d.apps, d.inputs.appName)
		if err != nil {
			return nil, fmt.Errorf("failed to find app: %w", err)
		}
		if app == nil {
			return nil, fmt.Errorf("app %q not found", d.inputs.appName)
		}
		return app.Spec, nil
	}

	specBytes, err := os.ReadFile(d.inputs.appSpecLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to read app spec at %q: %w", d.inputs.appSpecLocation, err)
	}

	expanded := utils.ExpandEnvRetainingBindables(string(specBytes))

	var spec godo.AppSpec
	if err := yaml.Unmarshal([]byte(expanded), &spec); err != nil {
		return nil, fmt.Errorf("failed to parse app spec: %w", err)
	}
	return &spec, nil
}

func (d *deployer) deploy(ctx context.Context, spec *godo.AppSpec) (*godo.App, error) {
	existingApp, err := utils.FindAppByName(ctx, d.apps, spec.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing app: %w", err)
	}

	var app *godo.App
	if existingApp == nil {
		d.action.Infof("Creating new app %q", spec.Name)
		app, _, err = d.apps.Create(ctx, &godo.AppCreateRequest{
			Spec:      spec,
			ProjectID: d.inputs.projectID,
		})
	} else {
		d.action.Infof("Updating existing app %q (%s)", spec.Name, existingApp.ID)
		app, _, err = d.apps.Update(ctx, existingApp.ID, &godo.AppUpdateRequest{Spec: spec})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create/update app: %w", err)
	}

	app, err = d.waitForDeployment(ctx, app)
	if err != nil {
		return app, err
	}
	return app, nil
}

func (d *deployer) waitForDeployment(ctx context.Context, app *godo.App) (*godo.App, error) {
	d.action.Infof("Waiting for deployment to complete...")
	for {
		deployments, _, err := d.apps.ListDeployments(ctx, app.ID, &godo.ListOptions{})
		if err != nil {
			return app, fmt.Errorf("failed to list deployments: %w", err)
		}
		if len(deployments) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		latest := deployments[0]
		switch latest.Phase {
		case godo.DeploymentPhase_Active:
			d.action.Infof("Deployment succeeded")
			app, _, err = d.apps.Get(ctx, app.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to refresh app after deployment: %w", err)
			}
			if d.inputs.printBuildLogs {
				d.printLogs(ctx, app.ID, latest.ID, godo.AppLogTypeBuild, "build")
			}
			if d.inputs.printDeployLogs {
				d.printLogs(ctx, app.ID, latest.ID, godo.AppLogTypeDeploy, "deploy")
			}
			return app, nil
		case godo.DeploymentPhase_Error, godo.DeploymentPhase_Canceled:
			d.printLogs(ctx, app.ID, latest.ID, godo.AppLogTypeBuild, "build")
			d.printLogs(ctx, app.ID, latest.ID, godo.AppLogTypeDeploy, "deploy")
			return app, fmt.Errorf("deployment phase: %s", latest.Phase)
		default:
			d.action.Infof("Deployment phase: %s — waiting...", latest.Phase)
			time.Sleep(5 * time.Second)
		}
	}
}

func (d *deployer) printLogs(ctx context.Context, appID, deployID string, logType godo.AppLogType, label string) {
	logs, _, err := d.apps.GetLogs(ctx, appID, deployID, "", logType, true, 0)
	if err != nil {
		d.action.Warningf("failed to get %s logs: %v", label, err)
		return
	}
	if logs.LiveURL == "" {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, logs.LiveURL, http.NoBody)
	if err != nil {
		d.action.Warningf("failed to build %s log request: %v", label, err)
		return
	}
	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.action.Warningf("failed to fetch %s log stream: %v", label, err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			d.action.Warningf("failed to close %s log response body: %v", label, err)
		}
	}()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		d.action.Warningf("failed to read %s logs: %v", label, err)
		return
	}
	d.action.SetOutput(label+"_logs", buf.String())
	fmt.Println(buf.String())
}
