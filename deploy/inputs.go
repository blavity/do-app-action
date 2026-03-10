package main

import (
	"github.com/blavity/do-app-action/utils"
	gha "github.com/sethvargo/go-githubactions"
)

// inputs are the inputs for the deploy action.
type inputs struct {
	token           string
	appSpecLocation string
	projectID       string
	appName         string
	printBuildLogs  bool
	printDeployLogs bool
	deployPRPreview bool
}

// getInputs gets the inputs for the action.
func getInputs(a *gha.Action) (inputs, error) {
	var in inputs
	for _, err := range []error{
		utils.InputAsString(a, "token", true, &in.token),
		utils.InputAsString(a, "app_spec_location", false, &in.appSpecLocation),
		utils.InputAsString(a, "project_id", false, &in.projectID),
		utils.InputAsString(a, "app_name", false, &in.appName),
		utils.InputAsBool(a, "print_build_logs", false, &in.printBuildLogs),
		utils.InputAsBool(a, "print_deploy_logs", false, &in.printDeployLogs),
		utils.InputAsBool(a, "deploy_pr_preview", false, &in.deployPRPreview),
	} {
		if err != nil {
			return in, err
		}
	}
	if in.appSpecLocation == "" {
		in.appSpecLocation = ".do/app.yaml"
	}
	return in, nil
}
