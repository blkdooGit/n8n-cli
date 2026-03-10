// Package integration contains integration tests for the n8n-cli
package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeValidateWorkflowFile writes a workflow JSON file for validation tests.
func writeValidateWorkflowFile(t *testing.T, dir, filename string, data map[string]interface{}) string {
	t.Helper()
	content, err := json.MarshalIndent(data, "", "  ")
	require.NoError(t, err)
	path := filepath.Join(dir, filename)
	require.NoError(t, os.WriteFile(path, content, 0644))
	return path
}

func TestValidateValidWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validate-valid-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	nodeName := "Start"
	nodeType := "n8n-nodes-base.start"
	workflow := map[string]interface{}{
		"name": "Valid Workflow",
		"nodes": []n8n.Node{
			{Name: &nodeName, Type: &nodeType},
		},
		"connections": map[string]interface{}{},
		"settings":    map[string]interface{}{},
	}
	filePath := writeValidateWorkflowFile(t, tmpDir, "valid.json", workflow)

	output, err := runCommand(t, "workflows", "validate", filePath)

	// A valid workflow should not return an error.
	assert.NoError(t, err)
	assert.Contains(t, output, "OK")
	assert.Contains(t, output, filePath)
}

func TestValidateInvalidWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validate-invalid-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Workflow with empty name — should trigger a validation error.
	workflow := map[string]interface{}{
		"name":        "",
		"nodes":       []interface{}{},
		"connections": map[string]interface{}{},
		"settings":    map[string]interface{}{},
	}
	filePath := writeValidateWorkflowFile(t, tmpDir, "invalid.json", workflow)

	output, err := runCommand(t, "workflows", "validate", filePath)

	// The command should return an error for invalid workflows.
	assert.Error(t, err)
	assert.Contains(t, output, "ERROR")
	assert.Contains(t, output, filePath)
}

func TestValidateDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validate-dir-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	nodeName1 := "Trigger"
	nodeType1 := "n8n-nodes-base.manualTrigger"
	wf1 := map[string]interface{}{
		"name": "Workflow One",
		"nodes": []n8n.Node{
			{Name: &nodeName1, Type: &nodeType1},
		},
		"connections": map[string]interface{}{},
		"settings":    map[string]interface{}{},
	}

	nodeName2 := "HTTP"
	nodeType2 := "n8n-nodes-base.httpRequest"
	wf2 := map[string]interface{}{
		"name": "Workflow Two",
		"nodes": []n8n.Node{
			{Name: &nodeName2, Type: &nodeType2},
		},
		"connections": map[string]interface{}{},
		"settings":    map[string]interface{}{},
	}

	writeValidateWorkflowFile(t, tmpDir, "wf1.json", wf1)
	writeValidateWorkflowFile(t, tmpDir, "wf2.json", wf2)

	output, err := runCommand(t, "workflows", "validate", "--directory", tmpDir)

	// Both workflows are valid — no error expected.
	assert.NoError(t, err)
	assert.Contains(t, output, "OK")
	// Both files should appear in the output.
	assert.Contains(t, output, "wf1.json")
	assert.Contains(t, output, "wf2.json")
}
