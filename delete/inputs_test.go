package main

import (
	"testing"

	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func TestGetInputs_AllFields(t *testing.T) {
	t.Setenv("INPUT_TOKEN", "mytoken")
	t.Setenv("INPUT_APP_NAME", "my-app")
	t.Setenv("INPUT_APP_ID", "")
	t.Setenv("INPUT_FROM_PR_PREVIEW", "false")
	t.Setenv("INPUT_IGNORE_NOT_FOUND", "true")

	a := gha.New()
	in, err := getInputs(a)
	assert.NoError(t, err)
	assert.Equal(t, "mytoken", in.token)
	assert.Equal(t, "my-app", in.appName)
	assert.False(t, in.fromPRPreview)
	assert.True(t, in.ignoreNotFound)
}

func TestGetInputs_Defaults(t *testing.T) {
	t.Setenv("INPUT_TOKEN", "tok")
	t.Setenv("INPUT_APP_ID", "abc-123")

	a := gha.New()
	in, err := getInputs(a)
	assert.NoError(t, err)
	assert.Equal(t, "abc-123", in.appID)
	assert.False(t, in.fromPRPreview)
	assert.False(t, in.ignoreNotFound)
}

func TestGetInputs_MissingToken(t *testing.T) {
	a := gha.New()
	_, err := getInputs(a)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}
