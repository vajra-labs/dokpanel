package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	app := SetupApp()

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Pong!", string(body))
}

func TestPong(t *testing.T) {
	app := SetupApp()

	req := httptest.NewRequest(http.MethodGet, "/api/pong", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Ping!", string(body))
}

func TestHealth(t *testing.T) {
	app := SetupApp()

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	// Check required fields exist
	assert.Contains(t, result, "uptime")
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "environment")
	assert.Contains(t, result, "timestamp")
	assert.Contains(t, result, "memory")

	assert.Equal(t, "v1.0.0", result["version"])
	assert.Equal(t, "test", result["environment"])
}

func TestNotFound(t *testing.T) {
	app := SetupApp()

	req := httptest.NewRequest(http.MethodGet, "/api/unknown-route", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, "Wrong Path", result["message"])
}
