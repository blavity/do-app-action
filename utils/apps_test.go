package utils_test

import (
	"context"
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
