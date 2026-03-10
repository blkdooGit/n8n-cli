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

func TestCreateProjectCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "create <NAME>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				name := args[0]

				if err := fakeClient.CreateProject(name); err != nil {
					return fmt.Errorf("failed to create project: %w", err)
				}
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "Project %q created successfully\n", name)
				return err
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	t.Run("successfully creates project", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.CreateProjectReturns(nil)

		cmd.SetArgs([]string{"My New Project"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), `Project "My New Project" created successfully`)
		assert.Equal(t, 1, fakeClient.CreateProjectCallCount())

		passedName := fakeClient.CreateProjectArgsForCall(0)
		assert.Equal(t, "My New Project", passedName)
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.CreateProjectReturns(errors.New("insufficient permissions"))

		cmd.SetArgs([]string{"Forbidden Project"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient permissions")
		assert.Equal(t, 1, fakeClient.CreateProjectCallCount())
	})
}
