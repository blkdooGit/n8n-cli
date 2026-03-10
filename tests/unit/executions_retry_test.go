package unit

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRetryExecutionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "retry <EXECUTION_ID>",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				id := args[0]
				loadWorkflow, _ := cmd.Flags().GetBool("load-workflow")

				execution, err := fakeClient.RetryExecution(id, loadWorkflow)
				if err != nil {
					return fmt.Errorf("failed to retry execution: %w", err)
				}

				status := ""
				if execution != nil && execution.Status != nil {
					status = string(*execution.Status)
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "Execution retried. New status: %s\n", status)
				return err
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().Bool("load-workflow", false, "Load the latest workflow version before retrying")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	t.Run("successfully retries execution", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		status := n8n.ExecutionStatusSuccess
		fakeClient.RetryExecutionReturns(&n8n.Execution{
			Status: &status,
		}, nil)

		cmd.SetArgs([]string{"exec-42"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Execution retried. New status: success")
		assert.Equal(t, 1, fakeClient.RetryExecutionCallCount())

		passedID, passedLoadWorkflow := fakeClient.RetryExecutionArgsForCall(0)
		assert.Equal(t, "exec-42", passedID)
		assert.False(t, passedLoadWorkflow)
	})

	t.Run("successfully retries with --load-workflow flag", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		status := n8n.ExecutionStatusRunning
		fakeClient.RetryExecutionReturns(&n8n.Execution{
			Status: &status,
		}, nil)

		cmd.SetArgs([]string{"exec-99"})
		err := cmd.Flags().Set("load-workflow", "true")
		assert.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Execution retried. New status: running")

		passedID, passedLoadWorkflow := fakeClient.RetryExecutionArgsForCall(0)
		assert.Equal(t, "exec-99", passedID)
		assert.True(t, passedLoadWorkflow)
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.RetryExecutionReturns(nil, errors.New("execution not found"))

		cmd.SetArgs([]string{"exec-missing"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution not found")
		assert.Equal(t, 1, fakeClient.RetryExecutionCallCount())
	})
}
