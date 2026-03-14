package utils_test

import (
	"testing"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

// ---------------------------------------------------------------------------
// PRRefFromContext tests
// ---------------------------------------------------------------------------

func TestPRRefFromContext_Valid(t *testing.T) {
	ghCtx := &gha.GitHubContext{
		Event: map[string]any{
			"pull_request": map[string]any{
				"number": float64(42),
			},
		},
	}
	ref, err := utils.PRRefFromContext(ghCtx)
	require.NoError(t, err)
	assert.Equal(t, "42/merge", ref)
}

func TestPRRefFromContext_MissingPRField(t *testing.T) {
	ghCtx := &gha.GitHubContext{
		Event: map[string]any{},
	}
	_, err := utils.PRRefFromContext(ghCtx)
	assert.ErrorContains(t, err, "pull_request field didn't exist")
}

func TestPRRefFromContext_MissingNumber(t *testing.T) {
	ghCtx := &gha.GitHubContext{
		Event: map[string]any{
			"pull_request": map[string]any{},
		},
	}
	_, err := utils.PRRefFromContext(ghCtx)
	assert.ErrorContains(t, err, "missing pull request number")
}

// ---------------------------------------------------------------------------
// SanitizeSpecForPullRequestPreview tests
// ---------------------------------------------------------------------------

func TestSanitizeSpecForPullRequestPreview_ClearsDomainAndAlerts(t *testing.T) {
	ghCtx := &gha.GitHubContext{
		ServerURL:       "https://github.com",
		Repository:      "myorg/myrepo",
		RepositoryOwner: "myorg",
		HeadRef:         "refs/pull/7/head",
		Event: map[string]any{
			"pull_request": map[string]any{
				"number": float64(7),
			},
		},
	}
	spec := &godo.AppSpec{
		Name: "original-name",
		Domains: []*godo.AppDomainSpec{
			{Domain: "example.com"},
		},
		Alerts: []*godo.AppAlertSpec{
			{Rule: godo.AppAlertSpecRule_DeploymentFailed},
		},
	}

	err := utils.SanitizeSpecForPullRequestPreview(spec, ghCtx)
	require.NoError(t, err)
	assert.Nil(t, spec.Domains, "domains should be cleared")
	assert.Nil(t, spec.Alerts, "alerts should be cleared")
	assert.NotEqual(t, "original-name", spec.Name, "name should be replaced")
	assert.LessOrEqual(t, len(spec.Name), 32)
}

func TestSanitizeSpecForPullRequestPreview_MissingPRNumber(t *testing.T) {
	ghCtx := &gha.GitHubContext{
		Event: map[string]any{},
	}
	spec := &godo.AppSpec{Name: "my-app"}
	err := utils.SanitizeSpecForPullRequestPreview(spec, ghCtx)
	assert.ErrorContains(t, err, "failed to get PR number")
}

func TestSanitizeSpecForPullRequestPreview_MutatesComponentBranch(t *testing.T) {
	ghCtx := &gha.GitHubContext{
		ServerURL:       "https://github.com",
		Repository:      "myorg/myrepo",
		RepositoryOwner: "myorg",
		HeadRef:         "refs/pull/9/head",
		Event: map[string]any{
			"pull_request": map[string]any{
				"number": float64(9),
			},
		},
	}

	spec := &godo.AppSpec{
		Name: "my-app",
		Services: []*godo.AppServiceSpec{
			{
				Name: "web",
				GitHub: &godo.GitHubSourceSpec{
					Repo:         "myorg/myrepo",
					Branch:       "main",
					DeployOnPush: true,
				},
			},
		},
	}

	err := utils.SanitizeSpecForPullRequestPreview(spec, ghCtx)
	require.NoError(t, err)
	assert.Equal(t, "refs/pull/9/head", spec.Services[0].GitHub.Branch)
	assert.False(t, spec.Services[0].GitHub.DeployOnPush)
}
