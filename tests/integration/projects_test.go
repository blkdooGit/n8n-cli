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

func TestListProjects(t *testing.T) {
	pType := "personal"
	projects := []n8n.Project{
		{Id: stringPtr("proj-1"), Name: "My Project", Type: &pType},
		{Id: stringPtr("proj-2"), Name: "Team Project", Type: stringPtr("team")},
	}
	projectList := n8n.ProjectList{Data: &projects}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/projects" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(projectList); err != nil {
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

	output, err := runCommand(t, "projects", "list")

	require.NoError(t, err)
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "TYPE")
	assert.Contains(t, output, "proj-1")
	assert.Contains(t, output, "My Project")
	assert.Contains(t, output, "proj-2")
	assert.Contains(t, output, "Team Project")
}

func TestCreateProject(t *testing.T) {
	createCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/projects" && r.Method == http.MethodPost {
			createCalled = true
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			resp := n8n.Project{
				Id:   stringPtr("new-proj-id"),
				Name: "New Integration Project",
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

	output, err := runCommand(t, "projects", "create", "New Integration Project")

	require.NoError(t, err)
	assert.True(t, createCalled, "Expected POST /api/v1/projects to be called")
	assert.Contains(t, output, "created")
}
