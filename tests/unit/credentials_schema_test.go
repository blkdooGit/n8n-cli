package unit

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetCredentialSchemaCommand(t *testing.T) {
	fakeClient := &clientfakes.FakeClientInterface{}
	var stdout, stderr bytes.Buffer

	setupTestCommand := func() *cobra.Command {
		cmd := &cobra.Command{
			Use:  "schema <TYPE>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				credType := args[0]

				schema, err := fakeClient.GetCredentialSchema(credType)
				if err != nil {
					return err
				}

				for k := range schema {
					cmd.Printf("property: %s\n", k)
				}
				return nil
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	t.Run("successfully retrieves credential schema", func(t *testing.T) {
		cmd := setupTestCommand()
		schema := map[string]interface{}{
			"host":     "string",
			"port":     "number",
			"username": "string",
			"password": "string",
		}
		fakeClient.GetCredentialSchemaReturns(schema, nil)

		cmd.SetArgs([]string{"postgres"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Equal(t, 1, fakeClient.GetCredentialSchemaCallCount())

		credType := fakeClient.GetCredentialSchemaArgsForCall(0)
		assert.Equal(t, "postgres", credType)

		output := stdout.String()
		assert.Contains(t, output, "host")
		assert.Contains(t, output, "port")
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetCredentialSchemaReturns(nil, errors.New("credential type not found"))

		cmd.SetArgs([]string{"unknownType"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credential type not found")
	})
}
