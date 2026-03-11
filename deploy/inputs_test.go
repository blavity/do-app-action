package main

import (
	"testing"

	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func TestGetInputs_AllFields(t *testing.T) {
	t.Setenv("INPUT_TOKEN", "mytoken")
	t.Setenv("INPUT_APP_SPEC_LOCATION", "infra/app.yaml")
	t.Setenv("INPUT_PROJECT_ID", "proj-123")
	t.Setenv("INPUT_APP_NAME", "my-app")
	t.Setenv("INPUT_PRINT_BUILD_LOGS", "true")
	t.Setenv("INPUT_PRINT_DEPLOY_LOGS", "true")
	t.Setenv("INPUT_DEPLOY_PR_PREVIEW", "true")

	a := gha.New()
	in, err := getInputs(a)
	assert.NoError(t, err)
	assert.Equal(t, "mytoken", in.token)
	assert.Equal(t, "infra/app.yaml", in.appSpecLocation)
	assert.Equal(t, "proj-123", in.projectID)
	assert.Equal(t, "my-app", in.appName)
	assert.True(t, in.printBuildLogs)
	assert.True(t, in.printDeployLogs)
	assert.True(t, in.deployPRPreview)
}

func TestGetInputs_Defaults(t *testing.T) {
	t.Setenv("INPUT_TOKEN", "tok")

	a := gha.New()
	in, err := getInputs(a)
	assert.NoError(t, err)
	// app_spec_location defaults to .do/app.yaml when not provided
	assert.Equal(t, ".do/app.yaml", in.appSpecLocation)
	assert.False(t, in.printBuildLogs)
	assert.False(t, in.printDeployLogs)
	assert.False(t, in.deployPRPreview)
}

func TestGetInputs_MissingToken(t *testing.T) {
	a := gha.New()
	_, err := getInputs(a)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}
