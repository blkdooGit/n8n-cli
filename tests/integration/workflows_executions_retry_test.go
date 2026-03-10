// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryExecution(t *testing.T) {
	status := n8n.ExecutionStatusSuccess
	retried := n8n.Execution{
		Id:     float32Ptr(1001),
		Status: &status,
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/executions/1001/retry" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(retried); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
	}))
	defer mockServer.Close()

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()
	defer viper.Reset()

	output, err := runCommand(t, "workflows", "retry", "1001")

	require.NoError(t, err)
	assert.Contains(t, output, "success")
}

func TestRetryExecutionNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/executions/9999/retry" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"message": "Execution not found"}`)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
	}))
	defer mockServer.Close()

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()
	defer viper.Reset()

	_, err := runCommand(t, "workflows", "retry", "9999")

	assert.Error(t, err)
}
