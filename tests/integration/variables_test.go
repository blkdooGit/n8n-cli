// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListVariables(t *testing.T) {
	varType := "string"
	vars := []n8n.Variable{
		{Id: stringPtr("1"), Key: "API_URL", Value: "https://example.com", Type: &varType},
		{Id: stringPtr("2"), Key: "TIMEOUT", Value: "30", Type: &varType},
	}
	varList := n8n.VariableList{Data: &vars}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/variables" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(varList); err != nil {
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

	output, err := runCommand(t, "variables", "list")

	require.NoError(t, err)
	assert.Contains(t, output, "KEY")
	assert.Contains(t, output, "VALUE")
	assert.Contains(t, output, "TYPE")
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "API_URL")
	assert.Contains(t, output, "https://example.com")
	assert.Contains(t, output, "TIMEOUT")
	assert.Contains(t, output, "30")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "2")
}

func TestSetVariableCreate(t *testing.T) {
	// Empty list — no existing variable with this key
	emptyVarList := n8n.VariableList{Data: &[]n8n.Variable{}}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		switch {
		case r.URL.Path == "/api/v1/variables" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(emptyVarList); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

		case r.URL.Path == "/api/v1/variables" && r.Method == http.MethodPost:
			var v n8n.Variable
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			v.Id = stringPtr("99")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(v)

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
		}
	}))
	defer mockServer.Close()

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()
	defer viper.Reset()

	output, err := runCommand(t, "variables", "set", "NEW_KEY", "new_value")

	require.NoError(t, err)
	assert.Contains(t, output, "created")
}

func TestSetVariableUpdate(t *testing.T) {
	varType := "string"
	existing := []n8n.Variable{
		{Id: stringPtr("42"), Key: "EXISTING_KEY", Value: "old_value", Type: &varType},
	}
	varList := n8n.VariableList{Data: &existing}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		switch {
		case r.URL.Path == "/api/v1/variables" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(varList); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

		case r.URL.Path == "/api/v1/variables/42" && r.Method == http.MethodPut:
			var v n8n.Variable
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(v)

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
		}
	}))
	defer mockServer.Close()

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()
	defer viper.Reset()

	output, err := runCommand(t, "variables", "set", "EXISTING_KEY", "new_value")

	require.NoError(t, err)
	assert.Contains(t, output, "updated")
}

func TestDeleteVariable(t *testing.T) {
	deleteCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/variables/55" && r.Method == http.MethodDelete {
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

	output, err := runCommand(t, "variables", "delete", "55")

	require.NoError(t, err)
	assert.True(t, deleteCalled, "Expected DELETE /api/v1/variables/55 to be called")
	assert.Contains(t, output, "deleted")
}

func TestExportVariablesJSON(t *testing.T) {
	varType := "string"
	vars := []n8n.Variable{
		{Id: stringPtr("1"), Key: "KEY_ONE", Value: "value_one", Type: &varType},
		{Id: stringPtr("2"), Key: "KEY_TWO", Value: "value_two", Type: &varType},
	}
	varList := n8n.VariableList{Data: &vars}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/variables" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(varList); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
	}))
	defer mockServer.Close()

	tmpDir, err := os.MkdirTemp("", "variables-export-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	outFile := filepath.Join(tmpDir, "variables.json")

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()
	defer viper.Reset()

	output, runErr := runCommand(t, "variables", "export", "--file", outFile)
	require.NoError(t, runErr)
	assert.Contains(t, output, "Exported")

	data, readErr := os.ReadFile(outFile)
	require.NoError(t, readErr)

	var exported []n8n.Variable
	require.NoError(t, json.Unmarshal(data, &exported))
	assert.Len(t, exported, 2)

	keys := []string{exported[0].Key, exported[1].Key}
	assert.Contains(t, keys, "KEY_ONE")
	assert.Contains(t, keys, "KEY_TWO")
}

func TestImportVariablesFromJSON(t *testing.T) {
	// No existing variables
	emptyVarList := n8n.VariableList{Data: &[]n8n.Variable{}}
	createdCount := 0

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		switch {
		case r.URL.Path == "/api/v1/variables" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(emptyVarList); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

		case r.URL.Path == "/api/v1/variables" && r.Method == http.MethodPost:
			createdCount++
			var v n8n.Variable
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			v.Id = stringPtr(fmt.Sprintf("%d", createdCount))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(v)

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
		}
	}))
	defer mockServer.Close()

	// Create a temp JSON file with two variables to import
	tmpDir, err := os.MkdirTemp("", "variables-import-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	varType := "string"
	importVars := []n8n.Variable{
		{Key: "IMPORT_A", Value: "alpha", Type: &varType},
		{Key: "IMPORT_B", Value: "beta", Type: &varType},
	}
	data, err := json.MarshalIndent(importVars, "", "  ")
	require.NoError(t, err)

	importFile := filepath.Join(tmpDir, "vars.json")
	require.NoError(t, os.WriteFile(importFile, data, 0644))

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()
	defer viper.Reset()

	output, runErr := runCommand(t, "variables", "import", "--file", importFile)
	require.NoError(t, runErr)
	assert.True(t,
		strings.Contains(output, "created") || strings.Contains(output, "Imported"),
		"Expected output to mention created or Imported, got: %s", output)
	assert.Equal(t, 2, createdCount, "Expected 2 POST requests to create variables")
}
