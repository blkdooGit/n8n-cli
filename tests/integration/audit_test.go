// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/config"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildAuditResponse() n8n.Audit {
	credsReport := map[string]interface{}{
		"risk1": "Some credential risk detail",
		"risk2": "Another credential issue",
	}
	nodesReport := map[string]interface{}{
		"nodeRisk1": "Dangerous node configuration",
	}
	return n8n.Audit{
		CredentialsRiskReport: &credsReport,
		NodesRiskReport:       &nodesReport,
	}
}

func setupAuditMockServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()
	auditReport := buildAuditResponse()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintln(w, `{"error": "Unauthorized"}`)
			return
		}

		if r.URL.Path == "/api/v1/audit" && r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(auditReport); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error": "Not found"}`)
	}))

	viper.Reset()
	viper.Set("api_key", "test-api-key")
	viper.Set("instance_url", mockServer.URL)
	config.Initialize()

	cleanup := func() {
		mockServer.Close()
		viper.Reset()
	}
	return mockServer, cleanup
}

func TestGenerateAuditTable(t *testing.T) {
	_, cleanup := setupAuditMockServer(t)
	defer cleanup()

	output, err := runCommand(t, "audit", "generate")

	require.NoError(t, err)
	assert.Contains(t, output, "SECTION")
	assert.Contains(t, output, "RISK COUNT")
	assert.Contains(t, output, "Credentials Risk Report")
	assert.Contains(t, output, "Nodes Risk Report")
}

func TestGenerateAuditJSON(t *testing.T) {
	_, cleanup := setupAuditMockServer(t)
	defer cleanup()

	output, err := runCommand(t, "audit", "generate", "--output", "json")

	require.NoError(t, err)
	// Verify the output contains expected JSON fields from the audit report.
	assert.Contains(t, output, "Credentials Risk Report")
	assert.Contains(t, output, "Nodes Risk Report")
	// The output should be parseable as JSON.
	var parsed n8n.Audit
	assert.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed), "Expected valid JSON audit output")
}
