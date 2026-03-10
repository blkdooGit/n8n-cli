package workflows

import (
	"fmt"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ExecutionsRetryCmd = &cobra.Command{
	Use:   "retry <EXECUTION_ID>",
	Short: "Retry a failed execution",
	Args:  cobra.ExactArgs(1),
	RunE:  retryExecution,
}

func init() {
	rootcmd.GetWorkflowsCmd().AddCommand(ExecutionsRetryCmd)
	ExecutionsRetryCmd.Flags().Bool("load-workflow", false, "Load the latest workflow version before retrying")
}

func retryExecution(cmd *cobra.Command, args []string) error {
	id := args[0]
	loadWorkflow, _ := cmd.Flags().GetBool("load-workflow")
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	client := n8n.NewClient(instanceURL, apiKey)
	execution, err := client.RetryExecution(id, loadWorkflow)
	if err != nil {
		return fmt.Errorf("failed to retry execution: %w", err)
	}

	status := ""
	if execution != nil && execution.Status != nil {
		status = string(*execution.Status)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Execution retried. New status: %s\n", status)
	return err
}
