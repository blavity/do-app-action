package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

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
// sleep is replaced with a no-op and now returns a fixed time to keep tests fast.
func newTestRunner(t *testing.T, m *mockAppsService, in inputs) *runner {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "gh-output-*")
	if err != nil {
		t.Fatalf("create temp output file: %v", err)
	}
	_ = f.Close()
	t.Setenv("GITHUB_OUTPUT", f.Name())
	a := gha.New()
	fixed := time.Now()
	return &runner{
		action:       a,
		apps:         m,
		inputs:       in,
		pollInterval: 0,
		now:          func() time.Time { return fixed },
		sleep:        func(time.Duration) {},
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

func TestRun_ByID_Success_NoWait(t *testing.T) {
	m := &mockAppsService{}
	spec := &godo.AppSpec{Name: "my-app"}
	app := &godo.App{ID: "app-123", Spec: spec}
	updated := &godo.App{ID: "app-123", Spec: spec}

	m.On("Get", mock.Anything, "app-123").Return(app, okResp(), nil)
	m.On("Update", mock.Anything, "app-123", mock.MatchedBy(func(req *godo.AppUpdateRequest) bool {
		return req.Spec.Maintenance != nil && !req.Spec.Maintenance.Archive
	})).Return(updated, okResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123", waitForLive: false})
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

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123"})
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
	assert.ErrorContains(t, err, "failed to unarchive app")
}

func TestRun_AppNotFound_ByName_Ignore(t *testing.T) {
	m := &mockAppsService{}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, emptyListResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appName: "gone", ignoreNotFound: true})
	err := r.run(context.Background())
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// waitForLiveURL tests
// ---------------------------------------------------------------------------

func TestWaitForLiveURL_ImmediatelyLive(t *testing.T) {
	m := &mockAppsService{}
	liveApp := &godo.App{ID: "app-123", LiveURL: "https://example.com"}
	m.On("Get", mock.Anything, "app-123").Return(liveApp, okResp(), nil)

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123", waitTimeout: 30})
	url := r.waitForLiveURL(context.Background(), "app-123")
	assert.Equal(t, "https://example.com", url)
}

func TestWaitForLiveURL_Timeout(t *testing.T) {
	m := &mockAppsService{}
	// App never has a live URL; deadline is always in the past.
	notLiveApp := &godo.App{ID: "app-123", LiveURL: ""}
	m.On("Get", mock.Anything, "app-123").Return(notLiveApp, okResp(), nil)

	fixed := time.Now()
	callCount := 0
	// now() advances past the deadline after the first poll.
	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123", waitTimeout: 1})
	r.now = func() time.Time {
		callCount++
		if callCount == 1 {
			// First call sets deadline — return fixed time so deadline is 1s in future.
			return fixed
		}
		// Second call (loop condition) returns past the deadline.
		return fixed.Add(2 * time.Second)
	}

	url := r.waitForLiveURL(context.Background(), "app-123")
	assert.Equal(t, "", url)
}

func TestWaitForLiveURL_GetError_ThenLive(t *testing.T) {
	m := &mockAppsService{}
	liveApp := &godo.App{ID: "app-123", LiveURL: "https://live.example.com"}

	// First call returns an error, second returns a live app.
	m.On("Get", mock.Anything, "app-123").Return((*godo.App)(nil), okResp(), fmt.Errorf("transient")).Once()
	m.On("Get", mock.Anything, "app-123").Return(liveApp, okResp(), nil).Once()

	r := newTestRunner(t, m, inputs{token: "tok", appID: "app-123", waitTimeout: 0})
	// With waitTimeout==0 the loop is indefinite; we need it to terminate after
	// the second poll succeeds. The mock is set up to return live on 2nd call.
	url := r.waitForLiveURL(context.Background(), "app-123")
	assert.Equal(t, "https://live.example.com", url)
	m.AssertExpectations(t)
}
