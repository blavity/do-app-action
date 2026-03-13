package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/digitalocean/godo"
	gha "github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *mockAppsService) Update(ctx context.Context, appID string, req *godo.AppUpdateRequest) (*godo.App, *godo.Response, error) {
	args := m.Called(ctx, appID, req)
	return args.Get(0).(*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

// newTestRunner returns a runner wired to the given mock and inputs, using a
// gha.Action that writes outputs to a temp file so SetOutput does not panic.
func newTestRunner(t *testing.T, m *mockAppsService, in inputs) *runner {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "gh-output-*")
	if err != nil {
		t.Fatalf("create temp output file: %v", err)
	}
	_ = f.Close()
	t.Setenv("GITHUB_OUTPUT", f.Name())
	a := gha.New()
	return &runner{
		action: a,
		apps:   m,
		inputs: in,
	}
}

func emptyListResp() *godo.Response {
	return &godo.Response{Links: &godo.Links{}}
}

func okResp() *godo.Response {
	return &godo.Response{Response: &http.Response{StatusCode: http.StatusOK}}
}

func notFoundResp() *godo.Response {
	return &godo.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}
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

// ---------------------------------------------------------------------------
// resolveAppID tests
// ---------------------------------------------------------------------------

func TestResolveAppID_Direct(t *testing.T) {
	m := &mockAppsService{}
	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123"})

	id, err := r.resolveAppID(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "app-123", id)
	m.AssertNotCalled(t, "List")
}

func TestResolveAppID_ByName_Found(t *testing.T) {
	m := &mockAppsService{}
	app := &godo.App{ID: "app-xyz", Spec: &godo.AppSpec{Name: "my-app"}}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{app}, emptyListResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appName: "my-app"})
	id, err := r.resolveAppID(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "app-xyz", id)
}

func TestResolveAppID_ByName_NotFound_Error(t *testing.T) {
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appName: "missing"})
	_, err := r.resolveAppID(context.Background())
	assert.ErrorContains(t, err, "missing")
}

func TestResolveAppID_ByName_NotFound_Ignore(t *testing.T) {
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appName: "missing", ignoreNotFound: true})
	id, err := r.resolveAppID(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "", id)
}

func TestResolveAppID_ListError(t *testing.T) {
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), fmt.Errorf("api error"))

	r := newTestRunner(t, m, inputs{token: "tok", appName: "my-app"})
	_, err := r.resolveAppID(context.Background())
	assert.ErrorContains(t, err, "api error")
}

func TestResolveAppID_FromPRPreview_NotFound_Ignore(t *testing.T) {
	setupPRPreviewEnv(t, 42, "myorg", "myrepo")

	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", fromPRPreview: true, ignoreNotFound: true})
	id, err := r.resolveAppID(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "", id)
}

func TestResolveAppID_FromPRPreview_NotFound_Error(t *testing.T) {
	setupPRPreviewEnv(t, 7, "acme", "site")

	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", fromPRPreview: true, ignoreNotFound: false})
	_, err := r.resolveAppID(context.Background())
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// run tests
// ---------------------------------------------------------------------------

func TestRun_NoIdentifier(t *testing.T) {
	m := &mockAppsService{}
	r := newTestRunner(t, m, inputs{token: "tok"})
	err := r.run(context.Background())
	assert.ErrorContains(t, err, "either app_id, app_name, or from_pr_preview")
}

func TestRun_ByID_Success(t *testing.T) {
	m := &mockAppsService{}
	spec := &godo.AppSpec{Name: "my-app"}
	app := &godo.App{ID: "app-123", Spec: spec}
	updated := &godo.App{ID: "app-123", Spec: spec}

	m.On("Get", mock.Anything, "app-123").Return(app, okResp(), nil)
	m.On("Update", mock.Anything, "app-123", mock.MatchedBy(func(req *godo.AppUpdateRequest) bool {
		return req.Spec.Maintenance != nil && req.Spec.Maintenance.Archive
	})).Return(updated, okResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123"})
	err := r.run(context.Background())
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestRun_GetNotFound_Ignore(t *testing.T) {
	m := &mockAppsService{}
	m.On("Get", mock.Anything, "app-123").Return((*godo.App)(nil), notFoundResp(), fmt.Errorf("not found"))

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123", ignoreNotFound: true})
	err := r.run(context.Background())
	assert.NoError(t, err)
}

func TestRun_GetNotFound_Error(t *testing.T) {
	m := &mockAppsService{}
	m.On("Get", mock.Anything, "app-123").Return((*godo.App)(nil), notFoundResp(), fmt.Errorf("not found"))

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123", ignoreNotFound: false})
	err := r.run(context.Background())
	assert.ErrorContains(t, err, "failed to get app")
}

func TestRun_UpdateError(t *testing.T) {
	m := &mockAppsService{}
	spec := &godo.AppSpec{Name: "my-app"}
	app := &godo.App{ID: "app-123", Spec: spec}

	m.On("Get", mock.Anything, "app-123").Return(app, okResp(), nil)
	m.On("Update", mock.Anything, "app-123", mock.Anything).Return((*godo.App)(nil), okResp(), fmt.Errorf("update failed"))

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123"})
	err := r.run(context.Background())
	assert.ErrorContains(t, err, "failed to archive app")
}

func TestRun_AppNotFound_ByName_Ignore(t *testing.T) {
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appName: "gone", ignoreNotFound: true})
	err := r.run(context.Background())
	assert.NoError(t, err)
}
