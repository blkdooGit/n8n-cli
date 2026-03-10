package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListVariablesCommand(t *testing.T) {
	fakeClient := &clientfakes.FakeClientInterface{}
	var stdout, stderr bytes.Buffer

	setupTestCommand := func() *cobra.Command {
		cmd := &cobra.Command{
			Use: "list",
			RunE: func(cmd *cobra.Command, args []string) error {
				outputFormat, _ := cmd.Flags().GetString("output")

				var allVariables []n8n.Variable
				var cursor string
				limit := 100
				for {
					varList, err := fakeClient.GetVariables(&limit, cursor)
					if err != nil {
						return err
					}
					if varList.Data != nil {
						allVariables = append(allVariables, *varList.Data...)
					}
					if varList.NextCursor == nil || *varList.NextCursor == "" {
						break
					}
					cursor = *varList.NextCursor
				}

				switch outputFormat {
				case "json":
					data, err := json.MarshalIndent(allVariables, "", "  ")
					if err != nil {
						return err
					}
					_, err = cmd.OutOrStdout().Write(append(data, '\n'))
					return err
				default:
					cmd.Println("KEY\tVALUE\tTYPE\tID")
					for _, v := range allVariables {
						id := ""
						if v.Id != nil {
							id = *v.Id
						}
						varType := ""
						if v.Type != nil {
							varType = *v.Type
						}
						cmd.Printf("%s\t%s\t%s\t%s\n", v.Key, v.Value, varType, id)
					}
					return nil
				}
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	createSampleVariableList := func(count int) *n8n.VariableList {
		variables := make([]n8n.Variable, count)
		for i := 0; i < count; i++ {
			id := stringPtr("var-id-" + string(rune('1'+i)))
			varType := stringPtr("string")
			variables[i] = n8n.Variable{
				Id:    id,
				Key:   "VAR_" + string(rune('A'+i)),
				Value: "value-" + string(rune('a'+i)),
				Type:  varType,
			}
		}
		return &n8n.VariableList{Data: &variables}
	}

	t.Run("successfully lists variables in table format", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetVariablesReturns(createSampleVariableList(3), nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "VAR_A")
		assert.Contains(t, stdout.String(), "VAR_B")
		assert.Contains(t, stdout.String(), "VAR_C")
	})

	t.Run("outputs JSON when --output json is specified", func(t *testing.T) {
		cmd := setupTestCommand()
		err := cmd.Flags().Set("output", "json")
		assert.NoError(t, err)

		fakeClient.GetVariablesReturns(createSampleVariableList(2), nil)

		err = cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), `"key"`)
		assert.Contains(t, stdout.String(), `"value"`)
		assert.Contains(t, stdout.String(), "VAR_A")
	})

	t.Run("shows empty output when no variables exist", func(t *testing.T) {
		cmd := setupTestCommand()
		emptyData := []n8n.Variable{}
		fakeClient.GetVariablesReturns(&n8n.VariableList{Data: &emptyData}, nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		// Table header is printed but no data rows
		assert.Contains(t, stdout.String(), "KEY")
		assert.NotContains(t, stdout.String(), "VAR_")
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetVariablesReturns(nil, errors.New("API error"))

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})
}
