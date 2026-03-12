package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/blavity/do-app-action/utils"
)

func TestGenerateAppName_Length(t *testing.T) {
	name := utils.GenerateAppName("some-very-long-org", "a-very-long-repository-name", "123/merge")
	assert.LessOrEqual(t, len(name), 32, "app name must be at most 32 characters")
}

func TestGenerateAppName_Deterministic(t *testing.T) {
	a := utils.GenerateAppName("org", "repo", "42/merge")
	b := utils.GenerateAppName("org", "repo", "42/merge")
	assert.Equal(t, a, b)
}

func TestGenerateAppName_Unique(t *testing.T) {
	a := utils.GenerateAppName("org", "repo", "42/merge")
	b := utils.GenerateAppName("org", "repo", "43/merge")
	assert.NotEqual(t, a, b)
}
