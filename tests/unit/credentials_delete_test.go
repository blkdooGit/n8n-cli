package unit

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCredentialCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "delete <ID>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id := args[0]

				if err := fakeClient.DeleteCredential(id); err != nil {
					return fmt.Errorf("failed to delete credential: %w", err)
				}
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "Credential %q deleted successfully\n", id)
				return err
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	t.Run("successfully deletes credential", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.DeleteCredentialReturns(nil)

		cmd.SetArgs([]string{"cred-123"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), `Credential "cred-123" deleted successfully`)
		assert.Equal(t, 1, fakeClient.DeleteCredentialCallCount())

		deletedID := fakeClient.DeleteCredentialArgsForCall(0)
		assert.Equal(t, "cred-123", deletedID)
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.DeleteCredentialReturns(errors.New("credential not found"))

		cmd.SetArgs([]string{"nonexistent-cred"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credential not found")
		assert.Equal(t, 1, fakeClient.DeleteCredentialCallCount())
	})
}
