package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCreateCredentialCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "create",
			RunE: func(cmd *cobra.Command, args []string) error {
				name, _ := cmd.Flags().GetString("name")
				credType, _ := cmd.Flags().GetString("type")
				dataStr, _ := cmd.Flags().GetString("data")

				var credData map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &credData); err != nil {
					return fmt.Errorf("invalid JSON in --data: %w", err)
				}

				cred := &n8n.Credential{
					Name: name,
					Type: credType,
					Data: &credData,
				}

				resp, err := fakeClient.CreateCredential(cred)
				if err != nil {
					return fmt.Errorf("failed to create credential: %w", err)
				}

				id := ""
				if resp != nil && resp.Id != nil {
					id = *resp.Id
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Credential %q created with ID: %s\n", name, id)
				return err
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("name", "n", "", "Credential name (required)")
		cmd.Flags().StringP("type", "t", "", "Credential type (required)")
		cmd.Flags().StringP("data", "d", "", "Credential data as JSON string (required)")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	t.Run("successfully creates credential with valid JSON data", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		credID := "cred-abc-123"
		fakeClient.CreateCredentialReturns(&n8n.CreateCredentialResponse{
			Id:   &credID,
			Name: "My HubSpot Key",
			Type: "hubspotApi",
		}, nil)

		cmd.SetArgs([]string{})
		err := cmd.Flags().Set("name", "My HubSpot Key")
		assert.NoError(t, err)
		err = cmd.Flags().Set("type", "hubspotApi")
		assert.NoError(t, err)
		err = cmd.Flags().Set("data", `{"apiKey":"secret-key-value"}`)
		assert.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), `Credential "My HubSpot Key" created with ID: cred-abc-123`)
		assert.Equal(t, 1, fakeClient.CreateCredentialCallCount())

		passedCred := fakeClient.CreateCredentialArgsForCall(0)
		assert.Equal(t, "My HubSpot Key", passedCred.Name)
		assert.Equal(t, "hubspotApi", passedCred.Type)
		assert.NotNil(t, passedCred.Data)
		assert.Equal(t, "secret-key-value", (*passedCred.Data)["apiKey"])
	})

	t.Run("returns error with invalid JSON in --data flag", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		err := cmd.Flags().Set("name", "Bad Cred")
		assert.NoError(t, err)
		err = cmd.Flags().Set("type", "someType")
		assert.NoError(t, err)
		err = cmd.Flags().Set("data", `not valid json`)
		assert.NoError(t, err)

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON in --data")
		assert.Equal(t, 0, fakeClient.CreateCredentialCallCount())
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.CreateCredentialReturns(nil, errors.New("quota exceeded"))

		err := cmd.Flags().Set("name", "My Cred")
		assert.NoError(t, err)
		err = cmd.Flags().Set("type", "httpBasicAuth")
		assert.NoError(t, err)
		err = cmd.Flags().Set("data", `{"user":"admin","password":"pass"}`)
		assert.NoError(t, err)

		err = cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
		assert.Equal(t, 1, fakeClient.CreateCredentialCallCount())
	})
}
