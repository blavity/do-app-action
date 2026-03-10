package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
)

// SanitizeSpecForPullRequestPreview modifies the given AppSpec to be suitable for a pull request preview.
// This includes:
//   - Setting a unique app name.
//   - Unsetting any domains.
//   - Unsetting any alerts.
//   - Setting the reference of all relevant components to point to the PR's ref.
func SanitizeSpecForPullRequestPreview(spec *godo.AppSpec, ghCtx *gha.GitHubContext) error {
	repoOwner, repo := ghCtx.Repo()
	prRef, err := PRRefFromContext(ghCtx)
	if err != nil {
		return fmt.Errorf("failed to get PR number: %w", err)
	}

	spec.Name = GenerateAppName(repoOwner, repo, prRef)
	spec.Domains = nil
	spec.Alerts = nil

	if err := godo.ForEachAppSpecComponent(spec, func(c godo.AppBuildableComponentSpec) error {
		ref := c.GetGitHub()
		if ref == nil || ref.Repo != fmt.Sprintf("%s/%s", repoOwner, repo) {
			return nil
		}
		ref.DeployOnPush = false
		ref.Branch = ghCtx.HeadRef
		return nil
	}); err != nil {
		return fmt.Errorf("failed to sanitize buildable components: %w", err)
	}
	return nil
}

// GenerateAppName generates a unique app name based on the repoOwner, repo, and ref.
func GenerateAppName(repoOwner, repo, ref string) string {
	baseName := fmt.Sprintf("%s-%s-%s", repoOwner, repo, ref)
	baseName = strings.ToLower(baseName)
	baseName = strings.NewReplacer(
		"/", "-",
		":", "",
		"_", "-",
		".", "-",
	).Replace(baseName)

	hasher := sha256.New()
	hasher.Write([]byte(baseName))
	suffix := "-" + hex.EncodeToString(hasher.Sum(nil))[:8]

	limit := 32 - len(suffix)
	if len(baseName) < limit {
		limit = len(baseName)
	}

	return baseName[:limit] + suffix
}

// PRRefFromContext extracts the PR number from the given GitHub context.
func PRRefFromContext(ghCtx *gha.GitHubContext) (string, error) {
	prFields, ok := ghCtx.Event["pull_request"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("pull_request field didn't exist on event: %v", ghCtx.Event)
	}
	prNumber, ok := prFields["number"].(float64)
	if !ok {
		return "", errors.New("missing pull request number")
	}
	return fmt.Sprintf("%d/merge", int(prNumber)), nil
}
