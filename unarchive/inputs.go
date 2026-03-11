package main

import (
	"github.com/blavity/do-app-action/utils"
	gha "github.com/sethvargo/go-githubactions"
)

// inputs are the inputs for the unarchive action.
type inputs struct {
	token          string
	appName        string
	appID          string
	fromPRPreview  bool
	ignoreNotFound bool
	waitForLive    bool
	waitTimeout    int
}

// getInputs gets the inputs for the action.
func getInputs(a *gha.Action) (inputs, error) {
	// Set defaults before parsing so that optional bool/int inputs with
	// non-zero defaults are preserved when the env var is absent.
	in := inputs{
		waitForLive: true,
		waitTimeout: 300,
	}
	for _, err := range []error{
		utils.InputAsString(a, "token", true, &in.token),
		utils.InputAsString(a, "app_name", false, &in.appName),
		utils.InputAsString(a, "app_id", false, &in.appID),
		utils.InputAsBool(a, "from_pr_preview", false, &in.fromPRPreview),
		utils.InputAsBool(a, "ignore_not_found", false, &in.ignoreNotFound),
		utils.InputAsBool(a, "wait_for_live", false, &in.waitForLive),
		utils.InputAsInt(a, "wait_timeout", false, &in.waitTimeout),
	} {
		if err != nil {
			return in, err
		}
	}
	return in, nil
}
