// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCredentialSchema(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"apiKey": map[string]interface{}{
				"type":        "string",
				"description": "HubSpot API Key",
			},
		},
		"required": []string{"apiKey"},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/credentials/schema/hubspotApi" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(schema); err != nil {
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

	output, err := runCommand(t, "credentials", "schema", "hubspotApi")

	require.NoError(t, err)
	// The command outputs JSON — verify it contains expected keys
	assert.Contains(t, output, "apiKey")
	assert.Contains(t, output, "properties")
}

func TestCreateCredential(t *testing.T) {
	now := time.Now()
	createCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/credentials" && r.Method == http.MethodPost {
			createCalled = true
			var cred n8n.Credential
			if err := json.NewDecoder(r.Body).Decode(&cred); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			resp := n8n.CreateCredentialResponse{
				Id:        stringPtr("cred-abc-123"),
				Name:      cred.Name,
				Type:      cred.Type,
				CreatedAt: &now,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
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

	output, err := runCommand(t, "credentials", "create",
		"--name", "My HubSpot Cred",
		"--type", "hubspotApi",
		"--data", `{"apiKey":"test-token-123"}`,
	)

	require.NoError(t, err)
	assert.True(t, createCalled, "Expected POST /api/v1/credentials to be called")
	assert.Contains(t, output, "created")
	assert.Contains(t, output, "cred-abc-123")
}

func TestDeleteCredential(t *testing.T) {
	deleteCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/credentials/cred-xyz-456" && r.Method == http.MethodDelete {
			deleteCalled = true
			w.WriteHeader(http.StatusOK)
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

	output, err := runCommand(t, "credentials", "delete", "cred-xyz-456")

	require.NoError(t, err)
	assert.True(t, deleteCalled, "Expected DELETE /api/v1/credentials/cred-xyz-456 to be called")
	assert.Contains(t, output, "deleted")
}
