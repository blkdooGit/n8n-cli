package unit

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetVariableCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Each sub-test creates its own fake so call counts never bleed across tests.
	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "set <KEY> <VALUE>",
			Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				key := args[0]
				value := args[1]
				varType, _ := cmd.Flags().GetString("type")

				limit := 250
				var cursor string
				var existingID *string
				for {
					varList, err := fakeClient.GetVariables(&limit, cursor)
					if err != nil {
						return err
					}
					if varList.Data != nil {
						for _, v := range *varList.Data {
							if v.Key == key {
								existingID = v.Id
								break
							}
						}
					}
					if existingID != nil || varList.NextCursor == nil || *varList.NextCursor == "" {
						break
					}
					cursor = *varList.NextCursor
				}

				variable := &n8n.Variable{
					Key:   key,
					Value: value,
					Type:  &varType,
				}

				if existingID != nil {
					if err := fakeClient.UpdateVariable(*existingID, variable); err != nil {
						return err
					}
					cmd.Printf("Variable %q updated successfully\n", key)
					return nil
				}

				if err := fakeClient.CreateVariable(variable); err != nil {
					return err
				}
				cmd.Printf("Variable %q created successfully\n", key)
				return nil
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("type", "t", "string", "Variable type")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	t.Run("creates new variable when key does not exist", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		emptyData := []n8n.Variable{}
		fakeClient.GetVariablesReturns(&n8n.VariableList{Data: &emptyData}, nil)
		fakeClient.CreateVariableReturns(nil)

		cmd.SetArgs([]string{"MY_KEY", "my_value"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), `Variable "MY_KEY" created successfully`)
		assert.Equal(t, 1, fakeClient.CreateVariableCallCount())
		assert.Equal(t, 0, fakeClient.UpdateVariableCallCount())

		createdVar := fakeClient.CreateVariableArgsForCall(0)
		assert.Equal(t, "MY_KEY", createdVar.Key)
		assert.Equal(t, "my_value", createdVar.Value)
	})

	t.Run("updates existing variable when key already exists", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		existingID := "existing-var-id"
		existingData := []n8n.Variable{
			{Id: &existingID, Key: "EXISTING_KEY", Value: "old_value", Type: stringPtr("string")},
		}
		fakeClient.GetVariablesReturns(&n8n.VariableList{Data: &existingData}, nil)
		fakeClient.UpdateVariableReturns(nil)

		cmd.SetArgs([]string{"EXISTING_KEY", "new_value"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), `Variable "EXISTING_KEY" updated successfully`)
		assert.Equal(t, 1, fakeClient.UpdateVariableCallCount())
		assert.Equal(t, 0, fakeClient.CreateVariableCallCount())

		updatedID, updatedVar := fakeClient.UpdateVariableArgsForCall(0)
		assert.Equal(t, "existing-var-id", updatedID)
		assert.Equal(t, "EXISTING_KEY", updatedVar.Key)
		assert.Equal(t, "new_value", updatedVar.Value)
	})

	t.Run("returns error when GetVariables API call fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GetVariablesReturns(nil, errors.New("connection refused"))

		cmd.SetArgs([]string{"MY_KEY", "my_value"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
		assert.Equal(t, 0, fakeClient.CreateVariableCallCount())
		assert.Equal(t, 0, fakeClient.UpdateVariableCallCount())
	})

	t.Run("returns error when CreateVariable API call fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		emptyData := []n8n.Variable{}
		fakeClient.GetVariablesReturns(&n8n.VariableList{Data: &emptyData}, nil)
		fakeClient.CreateVariableReturns(errors.New("quota exceeded"))

		cmd.SetArgs([]string{"NEW_KEY", "new_value"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
	})
}
