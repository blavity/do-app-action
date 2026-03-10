package main

import (
	"testing"

	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func TestGetInputs_Defaults(t *testing.T) {
	t.Setenv("INPUT_TOKEN", "mytoken")
	t.Setenv("INPUT_APP_NAME", "my-app")

	a := gha.New()
	in, err := getInputs(a)
	assert.NoError(t, err)
	assert.Equal(t, "mytoken", in.token)
	assert.Equal(t, "my-app", in.appName)
	assert.True(t, in.waitForLive)
	assert.Equal(t, 300, in.waitTimeout)
	assert.False(t, in.fromPRPreview)
	assert.False(t, in.ignoreNotFound)
}

func TestGetInputs_CustomWait(t *testing.T) {
	t.Setenv("INPUT_TOKEN", "tok")
	t.Setenv("INPUT_APP_ID", "abc-123")
	t.Setenv("INPUT_WAIT_FOR_LIVE", "false")
	t.Setenv("INPUT_WAIT_TIMEOUT", "60")

	a := gha.New()
	in, err := getInputs(a)
	assert.NoError(t, err)
	assert.Equal(t, "abc-123", in.appID)
	assert.False(t, in.waitForLive)
	assert.Equal(t, 60, in.waitTimeout)
}

func TestGetInputs_MissingToken(t *testing.T) {
	a := gha.New()
	_, err := getInputs(a)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}
