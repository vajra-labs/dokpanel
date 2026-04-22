package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoints(t *testing.T) {
	t.Run("should return pong on /api/ping", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
		resp, err := App.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Pong!", string(body))
	})

	t.Run("should return ping on /api/pong", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/pong", nil)
		resp, err := App.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Ping!", string(body))
	})

	t.Run("should return health status on /api/health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		resp, err := App.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)

		assert.IsType(t, "", result["uptime"])
		assert.IsType(t, "", result["version"])
		assert.IsType(t, "", result["environment"])
		assert.IsType(t, "", result["timestamp"])

		memory, ok := result["memory"].(map[string]any)
		assert.True(t, ok)

		for _, key := range []string{"alloc", "totalAlloc", "sys", "heapAlloc", "heapSys"} {
			field, ok := memory[key].(map[string]any)
			assert.True(t, ok, "missing memory field: %s", key)
			assert.IsType(t, float64(0), field["bytes"])
			assert.IsType(t, "", field["human"])
		}
	})
}

