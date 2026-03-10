package utils_test

import (
	"os"
	"testing"

	"github.com/blavity/do-app-action/utils"
	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func actionWithInputs(t *testing.T, inputs map[string]string) *gha.Action {
	t.Helper()
	for k, v := range inputs {
		t.Setenv("INPUT_"+k, v)
	}
	return gha.New()
}

func TestInputAsString_Required_Present(t *testing.T) {
	a := actionWithInputs(t, map[string]string{"TOKEN": "tok"})
	var s string
	err := utils.InputAsString(a, "token", true, &s)
	assert.NoError(t, err)
	assert.Equal(t, "tok", s)
}

func TestInputAsString_Required_Missing(t *testing.T) {
	os.Unsetenv("INPUT_TOKEN")
	a := gha.New()
	var s string
	err := utils.InputAsString(a, "token", true, &s)
	assert.Error(t, err)
}

func TestInputAsBool_True(t *testing.T) {
	a := actionWithInputs(t, map[string]string{"FLAG": "true"})
	var b bool
	err := utils.InputAsBool(a, "flag", false, &b)
	assert.NoError(t, err)
	assert.True(t, b)
}

func TestInputAsBool_Invalid(t *testing.T) {
	a := actionWithInputs(t, map[string]string{"FLAG": "notabool"})
	var b bool
	err := utils.InputAsBool(a, "flag", false, &b)
	assert.Error(t, err)
}

func TestInputAsInt_Valid(t *testing.T) {
	a := actionWithInputs(t, map[string]string{"TIMEOUT": "300"})
	var n int
	err := utils.InputAsInt(a, "timeout", false, &n)
	assert.NoError(t, err)
	assert.Equal(t, 300, n)
}
