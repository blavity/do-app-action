package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockAppsService implements godo.AppsService for testing.
type mockAppsService struct {
	mock.Mock
	godo.AppsService
}

func (m *mockAppsService) List(ctx context.Context, opts *godo.ListOptions) ([]*godo.App, *godo.Response, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockAppsService) Get(ctx context.Context, appID string) (*godo.App, *godo.Response, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).(*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockAppsService) Create(ctx context.Context, req *godo.AppCreateRequest) (*godo.App, *godo.Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockAppsService) Update(ctx context.Context, appID string, req *godo.AppUpdateRequest) (*godo.App, *godo.Response, error) {
	args := m.Called(ctx, appID, req)
	return args.Get(0).(*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockAppsService) ListDeployments(ctx context.Context, appID string, opts *godo.ListOptions) ([]*godo.Deployment, *godo.Response, error) {
	args := m.Called(ctx, appID, opts)
	return args.Get(0).([]*godo.Deployment), args.Get(1).(*godo.Response), args.Error(2)
}

func (m *mockAppsService) GetLogs(ctx context.Context, appID, deployID, component string, logType godo.AppLogType, follow bool, tailLines int) (*godo.AppLogs, *godo.Response, error) {
	args := m.Called(ctx, appID, deployID, component, logType, follow, tailLines)
	return args.Get(0).(*godo.AppLogs), args.Get(1).(*godo.Response), args.Error(2)
}

// newTestDeployer returns a deployer wired to the given mock and inputs, using a
// gha.Action that writes outputs to a temp file so SetOutput does not panic.
func newTestDeployer(t *testing.T, m *mockAppsService, in inputs) *deployer {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "gh-output-*")
	if err != nil {
		t.Fatalf("create temp output file: %v", err)
	}
	_ = f.Close()
	t.Setenv("GITHUB_OUTPUT", f.Name())
	a := gha.New()
	return &deployer{
		action:     a,
		apps:       m,
		httpClient: http.DefaultClient,
		inputs:     in,
		sleep:      func(time.Duration) {},
	}
}

func emptyListResp() *godo.Response {
	return &godo.Response{Links: &godo.Links{}}
}

func okResp() *godo.Response {
	return &godo.Response{Response: &http.Response{StatusCode: http.StatusOK}}
}

// ---------------------------------------------------------------------------
// createSpec tests
// ---------------------------------------------------------------------------

func TestCreateSpec_FromFile(t *testing.T) {
	specContent := `name: test-app
region: nyc
services:
  - name: web
    image:
      registry_type: DOCKER_HUB
      repository: nginx
      tag: latest
`
	f, err := os.CreateTemp(t.TempDir(), "app-*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(specContent)
	require.NoError(t, err)
	_ = f.Close()

	m := &mockAppsService{}
	d := newTestDeployer(t, m, inputs{token: "tok", appSpecLocation: f.Name()})

	spec, err := d.createSpec(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-app", spec.Name)
}

func TestCreateSpec_FromFile_Missing(t *testing.T) {
	m := &mockAppsService{}
	d := newTestDeployer(t, m, inputs{token: "tok", appSpecLocation: "/nonexistent/path.yaml"})

	_, err := d.createSpec(context.Background())
	assert.ErrorContains(t, err, "failed to read app spec")
}

func TestCreateSpec_FromFile_InvalidYAML(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "app-*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString("{{{{not valid yaml")
	require.NoError(t, err)
	_ = f.Close()

	m := &mockAppsService{}
	d := newTestDeployer(t, m, inputs{token: "tok", appSpecLocation: f.Name()})

	_, err = d.createSpec(context.Background())
	assert.ErrorContains(t, err, "failed to parse app spec")
}

func TestCreateSpec_ByAppName_Found(t *testing.T) {
	spec := &godo.AppSpec{Name: "my-app"}
	app := &godo.App{ID: "app-123", Spec: spec}
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{app}, emptyListResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok", appName: "my-app"})
	result, err := d.createSpec(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "my-app", result.Name)
}

func TestCreateSpec_ByAppName_NotFound(t *testing.T) {
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok", appName: "missing"})
	_, err := d.createSpec(context.Background())
	assert.ErrorContains(t, err, "not found")
}

// ---------------------------------------------------------------------------
// deploy tests
// ---------------------------------------------------------------------------

func TestDeploy_CreateNew(t *testing.T) {
	spec := &godo.AppSpec{Name: "new-app"}
	createdApp := &godo.App{ID: "app-new", Spec: spec}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}
	refreshedApp := &godo.App{ID: "app-new", Spec: spec}

	m := &mockAppsService{}
	// No existing app found.
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)
	m.On("Create", mock.Anything, mock.Anything).Return(createdApp, okResp(), nil)
	m.On("ListDeployments", mock.Anything, "app-new", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil)
	m.On("Get", mock.Anything, "app-new").Return(refreshedApp, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	result, err := d.deploy(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, "app-new", result.ID)
	m.AssertExpectations(t)
}

func TestDeploy_UpdateExisting(t *testing.T) {
	spec := &godo.AppSpec{Name: "existing-app"}
	existingApp := &godo.App{ID: "app-existing", Spec: spec}
	updatedApp := &godo.App{ID: "app-existing", Spec: spec}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}
	refreshedApp := &godo.App{ID: "app-existing", Spec: spec}

	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{existingApp}, emptyListResp(), nil)
	m.On("Update", mock.Anything, "app-existing", mock.Anything).Return(updatedApp, okResp(), nil)
	m.On("ListDeployments", mock.Anything, "app-existing", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil)
	m.On("Get", mock.Anything, "app-existing").Return(refreshedApp, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	result, err := d.deploy(context.Background(), spec)
	assert.NoError(t, err)
	assert.Equal(t, "app-existing", result.ID)
	m.AssertExpectations(t)
}

func TestDeploy_CreateError(t *testing.T) {
	spec := &godo.AppSpec{Name: "new-app"}
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)
	m.On("Create", mock.Anything, mock.Anything).Return((*godo.App)(nil), okResp(), fmt.Errorf("quota exceeded"))

	d := newTestDeployer(t, m, inputs{token: "tok"})
	_, err := d.deploy(context.Background(), spec)
	assert.ErrorContains(t, err, "failed to create/update app")
}

// ---------------------------------------------------------------------------
// waitForDeployment tests
// ---------------------------------------------------------------------------

func TestWaitForDeployment_EmptyThenActive(t *testing.T) {
	app := &godo.App{ID: "app-123", Spec: &godo.AppSpec{Name: "my-app"}}
	refreshedApp := &godo.App{ID: "app-123", Spec: app.Spec}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}

	m := &mockAppsService{}
	// First poll returns empty, second returns active.
	m.On("ListDeployments", mock.Anything, "app-123", mock.Anything).Return([]*godo.Deployment{}, okResp(), nil).Once()
	m.On("ListDeployments", mock.Anything, "app-123", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil).Once()
	m.On("Get", mock.Anything, "app-123").Return(refreshedApp, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	result, err := d.waitForDeployment(context.Background(), app)
	assert.NoError(t, err)
	assert.Equal(t, "app-123", result.ID)
	m.AssertExpectations(t)
}

func TestWaitForDeployment_ErrorPhase(t *testing.T) {
	app := &godo.App{ID: "app-123", Spec: &godo.AppSpec{Name: "my-app"}}
	errorDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Error}

	m := &mockAppsService{}
	m.On("ListDeployments", mock.Anything, "app-123", mock.Anything).Return([]*godo.Deployment{errorDeployment}, okResp(), nil)
	// GetLogs is called for build and deploy logs on failure.
	m.On("GetLogs", mock.Anything, "app-123", "dep-1", "", godo.AppLogTypeBuild, true, 0).Return(&godo.AppLogs{}, okResp(), nil)
	m.On("GetLogs", mock.Anything, "app-123", "dep-1", "", godo.AppLogTypeDeploy, true, 0).Return(&godo.AppLogs{}, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	_, err := d.waitForDeployment(context.Background(), app)
	assert.ErrorContains(t, err, "deployment phase")
}

func TestWaitForDeployment_CanceledPhase(t *testing.T) {
	app := &godo.App{ID: "app-123", Spec: &godo.AppSpec{Name: "my-app"}}
	canceledDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Canceled}

	m := &mockAppsService{}
	m.On("ListDeployments", mock.Anything, "app-123", mock.Anything).Return([]*godo.Deployment{canceledDeployment}, okResp(), nil)
	m.On("GetLogs", mock.Anything, "app-123", "dep-1", "", godo.AppLogTypeBuild, true, 0).Return(&godo.AppLogs{}, okResp(), nil)
	m.On("GetLogs", mock.Anything, "app-123", "dep-1", "", godo.AppLogTypeDeploy, true, 0).Return(&godo.AppLogs{}, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	_, err := d.waitForDeployment(context.Background(), app)
	assert.ErrorContains(t, err, "deployment phase")
}

func TestWaitForDeployment_ListError(t *testing.T) {
	app := &godo.App{ID: "app-123"}
	m := &mockAppsService{}
	m.On("ListDeployments", mock.Anything, "app-123", mock.Anything).Return([]*godo.Deployment{}, okResp(), fmt.Errorf("network error"))

	d := newTestDeployer(t, m, inputs{token: "tok"})
	_, err := d.waitForDeployment(context.Background(), app)
	assert.ErrorContains(t, err, "failed to list deployments")
}

func TestWaitForDeployment_Active_WithLogs(t *testing.T) {
	app := &godo.App{ID: "app-123", Spec: &godo.AppSpec{Name: "my-app"}}
	refreshedApp := &godo.App{ID: "app-123", Spec: app.Spec}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}

	// Serve fake log content from a test HTTP server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "build log line")
	}))
	defer srv.Close()

	m := &mockAppsService{}
	m.On("ListDeployments", mock.Anything, "app-123", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil)
	m.On("Get", mock.Anything, "app-123").Return(refreshedApp, okResp(), nil)
	m.On("GetLogs", mock.Anything, "app-123", "dep-1", "", godo.AppLogTypeBuild, true, 0).
		Return(&godo.AppLogs{LiveURL: srv.URL}, okResp(), nil)
	m.On("GetLogs", mock.Anything, "app-123", "dep-1", "", godo.AppLogTypeDeploy, true, 0).
		Return(&godo.AppLogs{LiveURL: srv.URL}, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok", printBuildLogs: true, printDeployLogs: true})
	result, err := d.waitForDeployment(context.Background(), app)
	assert.NoError(t, err)
	assert.Equal(t, "app-123", result.ID)
}

// ---------------------------------------------------------------------------
// printLogs tests
// ---------------------------------------------------------------------------

func TestPrintLogs_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "log output here")
	}))
	defer srv.Close()

	m := &mockAppsService{}
	m.On("GetLogs", mock.Anything, "app-1", "dep-1", "", godo.AppLogTypeBuild, true, 0).
		Return(&godo.AppLogs{LiveURL: srv.URL}, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	// Should not panic or error; output goes to stdout and SetOutput.
	d.printLogs(context.Background(), "app-1", "dep-1", godo.AppLogTypeBuild, "build")
	m.AssertExpectations(t)
}

func TestPrintLogs_NoLiveURL(t *testing.T) {
	m := &mockAppsService{}
	m.On("GetLogs", mock.Anything, "app-1", "dep-1", "", godo.AppLogTypeBuild, true, 0).
		Return(&godo.AppLogs{LiveURL: ""}, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	// Should return early without error.
	d.printLogs(context.Background(), "app-1", "dep-1", godo.AppLogTypeBuild, "build")
	m.AssertExpectations(t)
}

func TestPrintLogs_GetLogsError(t *testing.T) {
	m := &mockAppsService{}
	m.On("GetLogs", mock.Anything, "app-1", "dep-1", "", godo.AppLogTypeBuild, true, 0).
		Return((*godo.AppLogs)(nil), okResp(), fmt.Errorf("logs unavailable"))

	d := newTestDeployer(t, m, inputs{token: "tok"})
	// Should warn and not panic.
	d.printLogs(context.Background(), "app-1", "dep-1", godo.AppLogTypeBuild, "build")
}

// ---------------------------------------------------------------------------
// run (end-to-end wiring)
// ---------------------------------------------------------------------------

func TestRun_FromFile_CreateNew(t *testing.T) {
	specContent := "name: brand-new\nregion: nyc\n"
	f, err := os.CreateTemp(t.TempDir(), "app-*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(specContent)
	require.NoError(t, err)
	_ = f.Close()

	createdApp := &godo.App{ID: "app-new", Spec: &godo.AppSpec{Name: "brand-new"}}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}

	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)
	m.On("Create", mock.Anything, mock.Anything).Return(createdApp, okResp(), nil)
	m.On("ListDeployments", mock.Anything, "app-new", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil)
	m.On("Get", mock.Anything, "app-new").Return(createdApp, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok", appSpecLocation: f.Name()})
	err = d.run(context.Background())
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestRun_SpecFileEnvExpansion(t *testing.T) {
	t.Setenv("MY_REGION", "sfo")
	specContent := "name: env-app\nregion: ${MY_REGION}\n"
	f, err := os.CreateTemp(t.TempDir(), "app-*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(specContent)
	require.NoError(t, err)
	_ = f.Close()

	createdApp := &godo.App{ID: "app-env", Spec: &godo.AppSpec{Name: "env-app"}}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}

	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)
	m.On("Create", mock.Anything, mock.MatchedBy(func(req *godo.AppCreateRequest) bool {
		return req.Spec.Region == "sfo"
	})).Return(createdApp, okResp(), nil)
	m.On("ListDeployments", mock.Anything, "app-env", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil)
	m.On("Get", mock.Anything, "app-env").Return(createdApp, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok", appSpecLocation: f.Name()})
	err = d.run(context.Background())
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestWaitForDeployment_ActiveRefreshError(t *testing.T) {
	app := &godo.App{ID: "app-123", Spec: &godo.AppSpec{Name: "my-app"}}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}

	m := &mockAppsService{}
	m.On("ListDeployments", mock.Anything, "app-123", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil)
	m.On("Get", mock.Anything, "app-123").Return((*godo.App)(nil), okResp(), fmt.Errorf("refresh error"))

	d := newTestDeployer(t, m, inputs{token: "tok"})
	_, err := d.waitForDeployment(context.Background(), app)
	assert.ErrorContains(t, err, "failed to refresh app after deployment")
}

func TestPrintLogs_HTTPFetchError(t *testing.T) {
	// Point LiveURL at a server that immediately closes the connection.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hijack and close to simulate a broken connection.
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, _ := hj.Hijack()
		_ = conn.Close()
	}))
	defer srv.Close()

	m := &mockAppsService{}
	m.On("GetLogs", mock.Anything, "app-1", "dep-1", "", godo.AppLogTypeBuild, true, 0).
		Return(&godo.AppLogs{LiveURL: srv.URL}, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok"})
	// Should warn and not panic.
	d.printLogs(context.Background(), "app-1", "dep-1", godo.AppLogTypeBuild, "build")
}

// setupPRPreviewEnv configures env vars so that gha.Action.Context() returns a
// context with the given pull request number and repository.
func setupPRPreviewEnv(t *testing.T, prNumber int, repoOwner, repo string) {
	t.Helper()
	eventJSON := fmt.Sprintf(`{"pull_request":{"number":%d}}`, prNumber)
	f, err := os.CreateTemp(t.TempDir(), "gh-event-*.json")
	if err != nil {
		t.Fatalf("create temp event file: %v", err)
	}
	_, _ = f.WriteString(eventJSON)
	_ = f.Close()
	t.Setenv("GITHUB_EVENT_PATH", f.Name())
	t.Setenv("GITHUB_REPOSITORY", repoOwner+"/"+repo)
	t.Setenv("GITHUB_REPOSITORY_OWNER", repoOwner)
}

func TestRun_DeployPRPreview_Success(t *testing.T) {
	setupPRPreviewEnv(t, 42, "myorg", "myrepo")

	specContent := "name: original-name\nregion: nyc\n"
	f, err := os.CreateTemp(t.TempDir(), "app-*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(specContent)
	require.NoError(t, err)
	_ = f.Close()

	// The deployer will sanitize the name before calling FindAppByName, so
	// return empty list (no existing app) then create.
	createdApp := &godo.App{ID: "app-pr", Spec: &godo.AppSpec{Name: "sanitized-pr-name"}}
	activeDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Active}

	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)
	m.On("Create", mock.Anything, mock.Anything).Return(createdApp, okResp(), nil)
	m.On("ListDeployments", mock.Anything, "app-pr", mock.Anything).Return([]*godo.Deployment{activeDeployment}, okResp(), nil)
	m.On("Get", mock.Anything, "app-pr").Return(createdApp, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok", appSpecLocation: f.Name(), deployPRPreview: true})
	err = d.run(context.Background())
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestRun_DeployError_StillSetsOutput(t *testing.T) {
	specContent := "name: failing-app\nregion: nyc\n"
	f, err := os.CreateTemp(t.TempDir(), "app-*.yaml")
	require.NoError(t, err)
	_, err = f.WriteString(specContent)
	require.NoError(t, err)
	_ = f.Close()

	createdApp := &godo.App{ID: "app-fail", Spec: &godo.AppSpec{Name: "failing-app"}}
	errorDeployment := &godo.Deployment{ID: "dep-1", Phase: godo.DeploymentPhase_Error}

	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)
	m.On("Create", mock.Anything, mock.Anything).Return(createdApp, okResp(), nil)
	m.On("ListDeployments", mock.Anything, "app-fail", mock.Anything).Return([]*godo.Deployment{errorDeployment}, okResp(), nil)
	m.On("GetLogs", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&godo.AppLogs{}, okResp(), nil)

	d := newTestDeployer(t, m, inputs{token: "tok", appSpecLocation: f.Name()})
	err = d.run(context.Background())
	// run returns the deploy error; app output is set because app is non-nil.
	assert.ErrorContains(t, err, "deployment phase")

	// Verify the app output was written to GITHUB_OUTPUT despite the deploy error.
	// gha.Action.SetOutput writes the value using heredoc format:
	//   key<<_GitHubActionsFileCommandDelimeter_\n<value>\n_GitHubActionsFileCommandDelimeter_\n
	outputContent, readErr := os.ReadFile(os.Getenv("GITHUB_OUTPUT"))
	assert.NoError(t, readErr)
	assert.Contains(t, string(outputContent), "app-fail", "expected app ID to be present in output")
}
