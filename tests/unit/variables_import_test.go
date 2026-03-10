package unit

import (
	"bufio"
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

func TestImportVariablesCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "import",
			RunE: func(cmd *cobra.Command, args []string) error {
				filePath, _ := cmd.Flags().GetString("file")

				data, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}

				var variables []n8n.Variable
				ext := strings.ToLower(filepath.Ext(filePath))
				switch ext {
				case ".yaml", ".yml":
					if err := yaml.Unmarshal(data, &variables); err != nil {
						return fmt.Errorf("failed to parse YAML: %w", err)
					}
				case ".env":
					scanner := bufio.NewScanner(strings.NewReader(string(data)))
					for scanner.Scan() {
						line := strings.TrimSpace(scanner.Text())
						if line == "" || strings.HasPrefix(line, "#") {
							continue
						}
						parts := strings.SplitN(line, "=", 2)
						if len(parts) != 2 {
							continue
						}
						variables = append(variables, n8n.Variable{
							Key:   strings.TrimSpace(parts[0]),
							Value: strings.TrimSpace(parts[1]),
						})
					}
				default:
					if err := json.Unmarshal(data, &variables); err != nil {
						return fmt.Errorf("failed to parse JSON: %w", err)
					}
				}

				// Fetch existing variables for upsert
				limit := 250
				existingMap := make(map[string]*n8n.Variable)
				var cursor string
				for {
					varList, err := fakeClient.GetVariables(&limit, cursor)
					if err != nil {
						return fmt.Errorf("failed to fetch existing variables: %w", err)
					}
					if varList.Data != nil {
						for i := range *varList.Data {
							v := &(*varList.Data)[i]
							existingMap[v.Key] = v
						}
					}
					if varList.NextCursor == nil || *varList.NextCursor == "" {
						break
					}
					cursor = *varList.NextCursor
				}

				created, updated := 0, 0
				for i := range variables {
					v := &variables[i]
					if existing, found := existingMap[v.Key]; found {
						if err := fakeClient.UpdateVariable(*existing.Id, v); err != nil {
							fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update %q: %v\n", v.Key, err)
							continue
						}
						updated++
					} else {
						if err := fakeClient.CreateVariable(v); err != nil {
							fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to create %q: %v\n", v.Key, err)
							continue
						}
						created++
					}
				}

				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Imported variables: %d created, %d updated\n", created, updated)
				return err
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("file", "f", "", "Input file path (required)")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	t.Run("successfully imports from JSON file (creates variables)", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		// No existing variables
		emptyData := []n8n.Variable{}
		fakeClient.GetVariablesReturns(&n8n.VariableList{Data: &emptyData}, nil)
		fakeClient.CreateVariableReturns(nil)

		vars := []n8n.Variable{
			{Key: "API_URL", Value: "https://example.com"},
			{Key: "TIMEOUT", Value: "30"},
		}
		jsonData, err := json.Marshal(vars)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		inFile := filepath.Join(tmpDir, "vars.json")
		err = os.WriteFile(inFile, jsonData, 0600)
		require.NoError(t, err)

		err = cmd.Flags().Set("file", inFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "2 created")
		assert.Contains(t, stdout.String(), "0 updated")
		assert.Equal(t, 2, fakeClient.CreateVariableCallCount())
		assert.Equal(t, 0, fakeClient.UpdateVariableCallCount())
	})

	t.Run("successfully imports from dotenv file", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		emptyData := []n8n.Variable{}
		fakeClient.GetVariablesReturns(&n8n.VariableList{Data: &emptyData}, nil)
		fakeClient.CreateVariableReturns(nil)

		dotenvContent := "# comment\nDB_HOST=localhost\nDB_PORT=5432\n"
		tmpDir := t.TempDir()
		inFile := filepath.Join(tmpDir, "vars.env")
		err := os.WriteFile(inFile, []byte(dotenvContent), 0600)
		require.NoError(t, err)

		err = cmd.Flags().Set("file", inFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "2 created")
		assert.Equal(t, 2, fakeClient.CreateVariableCallCount())

		firstVar := fakeClient.CreateVariableArgsForCall(0)
		assert.Equal(t, "DB_HOST", firstVar.Key)
		assert.Equal(t, "localhost", firstVar.Value)
	})

	t.Run("updates existing variables (upsert)", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		existingID := "existing-id-1"
		existingData := []n8n.Variable{
			{Id: &existingID, Key: "API_URL", Value: "https://old.example.com"},
		}
		fakeClient.GetVariablesReturns(&n8n.VariableList{Data: &existingData}, nil)
		fakeClient.UpdateVariableReturns(nil)
		fakeClient.CreateVariableReturns(nil)

		vars := []n8n.Variable{
			{Key: "API_URL", Value: "https://new.example.com"},
			{Key: "NEW_VAR", Value: "brand-new"},
		}
		jsonData, err := json.Marshal(vars)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		inFile := filepath.Join(tmpDir, "vars.json")
		err = os.WriteFile(inFile, jsonData, 0600)
		require.NoError(t, err)

		err = cmd.Flags().Set("file", inFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "1 created")
		assert.Contains(t, stdout.String(), "1 updated")
		assert.Equal(t, 1, fakeClient.UpdateVariableCallCount())
		assert.Equal(t, 1, fakeClient.CreateVariableCallCount())

		updatedID, updatedVar := fakeClient.UpdateVariableArgsForCall(0)
		assert.Equal(t, "existing-id-1", updatedID)
		assert.Equal(t, "API_URL", updatedVar.Key)
		assert.Equal(t, "https://new.example.com", updatedVar.Value)
	})

	t.Run("returns error when file not found", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		err := cmd.Flags().Set("file", "/nonexistent/path/vars.json")
		require.NoError(t, err)

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
		assert.Equal(t, 0, fakeClient.GetVariablesCallCount())
	})

	t.Run("returns error when GetVariables API fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(nil, errors.New("API unavailable"))

		vars := []n8n.Variable{{Key: "SOME_KEY", Value: "some_value"}}
		jsonData, err := json.Marshal(vars)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		inFile := filepath.Join(tmpDir, "vars.json")
		err = os.WriteFile(inFile, jsonData, 0600)
		require.NoError(t, err)

		err = cmd.Flags().Set("file", inFile)
		require.NoError(t, err)

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API unavailable")
		assert.Equal(t, 0, fakeClient.CreateVariableCallCount())
		assert.Equal(t, 0, fakeClient.UpdateVariableCallCount())
	})
}
