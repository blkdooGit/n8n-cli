package unit

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/edenreich/n8n-cli/cmd/workflows"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestValidateWorkflowsCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func() *cobra.Command {
		cmd := &cobra.Command{
			Use:  "validate [FILES...]",
			RunE: func(cmd *cobra.Command, args []string) error {
				return workflows.ValidateCmd.RunE(cmd, args)
			},
		}
		// Mirror flags from the real ValidateCmd
		cmd.Flags().StringP("directory", "d", "", "Validate all workflow files in a directory")
		cmd.Flags().Bool("strict", false, "Fail on warnings in addition to errors")

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	validWorkflowJSON := `{
  "name": "Test Workflow",
  "nodes": [
    {"id": "1", "name": "Start", "type": "n8n-nodes-base.start", "parameters": {}, "position": [0, 0]},
    {"id": "2", "name": "HTTP Request", "type": "n8n-nodes-base.httpRequest", "parameters": {}, "position": [200, 0]}
  ],
  "connections": {},
  "settings": {}
}`

	t.Run("valid workflow passes validation", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "validate-test-*")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		filePath := filepath.Join(tmpDir, "valid.json")
		err = os.WriteFile(filePath, []byte(validWorkflowJSON), 0644)
		assert.NoError(t, err)

		cmd := setupTestCommand()
		cmd.SetArgs([]string{filePath})
		err = cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "OK")
	})

	t.Run("workflow with empty name fails validation", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "validate-test-*")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		noNameJSON := `{
  "name": "",
  "nodes": [
    {"id": "1", "name": "Start", "type": "n8n-nodes-base.start", "parameters": {}, "position": [0, 0]}
  ],
  "connections": {},
  "settings": {}
}`
		filePath := filepath.Join(tmpDir, "no_name.json")
		err = os.WriteFile(filePath, []byte(noNameJSON), 0644)
		assert.NoError(t, err)

		cmd := setupTestCommand()
		cmd.SetArgs([]string{filePath})
		err = cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, stdout.String(), "ERROR")
		assert.Contains(t, stdout.String(), "workflow name is empty")
	})

	t.Run("workflow with no nodes fails validation", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "validate-test-*")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		noNodesJSON := `{
  "name": "Empty Workflow",
  "nodes": [],
  "connections": {},
  "settings": {}
}`
		filePath := filepath.Join(tmpDir, "no_nodes.json")
		err = os.WriteFile(filePath, []byte(noNodesJSON), 0644)
		assert.NoError(t, err)

		cmd := setupTestCommand()
		cmd.SetArgs([]string{filePath})
		err = cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, stdout.String(), "ERROR")
		assert.Contains(t, stdout.String(), "workflow has no nodes")
	})

	t.Run("workflow with duplicate node names fails validation", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "validate-test-*")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tmpDir) })

		duplicateNamesJSON := `{
  "name": "Duplicate Names Workflow",
  "nodes": [
    {"id": "1", "name": "HTTP Request", "type": "n8n-nodes-base.httpRequest", "parameters": {}, "position": [0, 0]},
    {"id": "2", "name": "HTTP Request", "type": "n8n-nodes-base.httpRequest", "parameters": {}, "position": [200, 0]}
  ],
  "connections": {},
  "settings": {}
}`
		filePath := filepath.Join(tmpDir, "duplicate_names.json")
		err = os.WriteFile(filePath, []byte(duplicateNamesJSON), 0644)
		assert.NoError(t, err)

		cmd := setupTestCommand()
		cmd.SetArgs([]string{filePath})
		err = cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, stdout.String(), "ERROR")
		assert.Contains(t, stdout.String(), "duplicate node name")
		assert.Contains(t, stdout.String(), "HTTP Request")
	})

	t.Run("non-existent file fails validation", func(t *testing.T) {
		cmd := setupTestCommand()
		cmd.SetArgs([]string{"/nonexistent/path/workflow.json"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, stdout.String(), "ERROR")
	})

	t.Run("no files provided returns error", func(t *testing.T) {
		cmd := setupTestCommand()
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no files to validate")
	})
}
