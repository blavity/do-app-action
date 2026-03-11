package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandEnvRetainingBindables_NormalSubstitution(t *testing.T) {
	t.Setenv("MY_VAR", "hello")
	result := ExpandEnvRetainingBindables("${MY_VAR} world")
	assert.Equal(t, "hello world", result)
}

func TestExpandEnvRetainingBindables_BindableWithDotRetained(t *testing.T) {
	// Variables with dots are bindable (DO app platform syntax) and must not be expanded.
	result := ExpandEnvRetainingBindables("${db.DATABASE_URL}")
	assert.Equal(t, "${db.DATABASE_URL}", result)
}

func TestExpandEnvRetainingBindables_AppWideVariablesRetained(t *testing.T) {
	// APP_DOMAIN, APP_URL, APP_ID are app-wide variables and must not be expanded
	// even when unset (they are resolved by the DO platform at runtime).
	for _, v := range []string{"APP_DOMAIN", "APP_URL", "APP_ID"} {
		result := ExpandEnvRetainingBindables("${" + v + "}")
		assert.Equal(t, "${"+v+"}", result, "expected %s to be retained", v)
	}
}

func TestExpandEnvRetainingBindables_UnsetNonBindableExpandsToEmpty(t *testing.T) {
	// An unset variable that is not an app-wide or bindable var expands to empty string.
	result := ExpandEnvRetainingBindables("${DEFINITELY_NOT_SET_XYZ}")
	assert.Equal(t, "", result)
}

func TestExpandEnvRetainingBindables_MixedInputs(t *testing.T) {
	t.Setenv("BUILD_TAG", "v1.2.3")
	result := ExpandEnvRetainingBindables("tag=${BUILD_TAG} url=${APP_URL} db=${db.DATABASE_URL}")
	assert.Equal(t, "tag=v1.2.3 url=${APP_URL} db=${db.DATABASE_URL}", result)
}
