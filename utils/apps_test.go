package utils_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/blavity/do-app-action/utils"
)

type mockAppsService struct {
	mock.Mock
	godo.AppsService
}

func (m *mockAppsService) List(ctx context.Context, opts *godo.ListOptions) ([]*godo.App, *godo.Response, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*godo.App), args.Get(1).(*godo.Response), args.Error(2)
}

func TestFindAppByName_Found(t *testing.T) {
	m := &mockAppsService{}
	app := &godo.App{ID: "abc", Spec: &godo.AppSpec{Name: "my-app"}}
	resp := &godo.Response{Links: &godo.Links{}}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{app}, resp, nil)

	result, err := utils.FindAppByName(context.Background(), m, "my-app")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc", result.ID)
}

func TestFindAppByName_NotFound(t *testing.T) {
	m := &mockAppsService{}
	resp := &godo.Response{Links: &godo.Links{}}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, resp, nil)

	result, err := utils.FindAppByName(context.Background(), m, "missing-app")
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestFindAppByName_MultiPage(t *testing.T) {
	m := &mockAppsService{}
	page1App := &godo.App{ID: "p1", Spec: &godo.AppSpec{Name: "other-app"}}
	page2App := &godo.App{ID: "p2", Spec: &godo.AppSpec{Name: "target-app"}}

	page1Resp := &godo.Response{
		Links: &godo.Links{Pages: &godo.Pages{Next: "http://example.com/apps?page=2"}},
	}
	page2Resp := &godo.Response{Links: &godo.Links{}}

	m.On("List", mock.Anything, &godo.ListOptions{}).Return([]*godo.App{page1App}, page1Resp, nil)
	m.On("List", mock.Anything, &godo.ListOptions{Page: 2}).Return([]*godo.App{page2App}, page2Resp, nil)

	result, err := utils.FindAppByName(context.Background(), m, "target-app")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "p2", result.ID)
	m.AssertExpectations(t)
}

func TestFindAppByName_ListError(t *testing.T) {
	m := &mockAppsService{}
	resp := &godo.Response{Links: &godo.Links{}}
	m.On("List", mock.Anything, mock.Anything).Return([]*godo.App{}, resp, fmt.Errorf("api unavailable"))

	_, err := utils.FindAppByName(context.Background(), m, "my-app")
	assert.ErrorContains(t, err, "failed to list apps")
}
