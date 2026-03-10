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
