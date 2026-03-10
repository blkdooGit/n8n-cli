package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestExportVariablesCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "export",
			RunE: func(cmd *cobra.Command, args []string) error {
				filePath, _ := cmd.Flags().GetString("file")
				format, _ := cmd.Flags().GetString("format")

				if format == "" {
					ext := strings.ToLower(filepath.Ext(filePath))
					switch ext {
					case ".yaml", ".yml":
						format = "yaml"
					case ".env":
						format = "dotenv"
					default:
						format = "json"
					}
				}

				var allVariables []n8n.Variable
				limit := 250
				var cursor string
				for {
					varList, err := fakeClient.GetVariables(&limit, cursor)
					if err != nil {
						return fmt.Errorf("failed to fetch variables: %w", err)
					}
					if varList.Data != nil {
						allVariables = append(allVariables, *varList.Data...)
					}
					if varList.NextCursor == nil || *varList.NextCursor == "" {
						break
					}
					cursor = *varList.NextCursor
				}

				var content []byte
				var err error
				switch format {
				case "yaml":
					content, err = yaml.Marshal(allVariables)
				case "dotenv":
					var sb strings.Builder
					for _, v := range allVariables {
						sb.WriteString(fmt.Sprintf("%s=%s\n", v.Key, v.Value))
					}
					content = []byte(sb.String())
				default:
					content, err = json.MarshalIndent(allVariables, "", "  ")
				}
				if err != nil {
					return fmt.Errorf("failed to serialize variables: %w", err)
				}

				if err := os.WriteFile(filePath, content, 0600); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}

				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Exported %d variables to %s\n", len(allVariables), filePath)
				return err
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("file", "f", "", "Output file path (required)")
		cmd.Flags().StringP("format", "", "", "Output format: json, yaml, dotenv")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	createSampleVariableList := func() *n8n.VariableList {
		varType := stringPtr("string")
		vars := []n8n.Variable{
			{Id: stringPtr("id-1"), Key: "DB_HOST", Value: "localhost", Type: varType},
			{Id: stringPtr("id-2"), Key: "DB_PORT", Value: "5432", Type: varType},
		}
		return &n8n.VariableList{Data: &vars}
	}

	t.Run("successfully exports to JSON file", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(createSampleVariableList(), nil)

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "vars.json")

		err := cmd.Flags().Set("file", outFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Exported 2 variables to")

		content, err := os.ReadFile(outFile)
		require.NoError(t, err)

		var result []n8n.Variable
		err = json.Unmarshal(content, &result)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "DB_HOST", result[0].Key)
		assert.Equal(t, "DB_PORT", result[1].Key)
	})

	t.Run("successfully exports to YAML file", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(createSampleVariableList(), nil)

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "vars.yaml")

		err := cmd.Flags().Set("file", outFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)

		content, err := os.ReadFile(outFile)
		require.NoError(t, err)

		var result []n8n.Variable
		err = yaml.Unmarshal(content, &result)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "DB_HOST", result[0].Key)
	})

	t.Run("successfully exports to dotenv file", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(createSampleVariableList(), nil)

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "vars.env")

		err := cmd.Flags().Set("file", outFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)

		content, err := os.ReadFile(outFile)
		require.NoError(t, err)

		lines := string(content)
		assert.Contains(t, lines, "DB_HOST=localhost")
		assert.Contains(t, lines, "DB_PORT=5432")
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(nil, errors.New("connection refused"))

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "vars.json")

		err := cmd.Flags().Set("file", outFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("auto-detects JSON format from .json extension", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(createSampleVariableList(), nil)

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "output.json")

		err := cmd.Flags().Set("file", outFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)

		content, err := os.ReadFile(outFile)
		require.NoError(t, err)

		// Valid JSON means format was auto-detected as json
		var result []n8n.Variable
		err = json.Unmarshal(content, &result)
		assert.NoError(t, err)
	})

	t.Run("auto-detects YAML format from .yml extension", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(createSampleVariableList(), nil)

		tmpDir := t.TempDir()
		outFile := filepath.Join(tmpDir, "output.yml")

		err := cmd.Flags().Set("file", outFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)

		content, err := os.ReadFile(outFile)
		require.NoError(t, err)

		var result []n8n.Variable
		err = yaml.Unmarshal(content, &result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}
