package variables

import (
	"fmt"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete <ID>",
	Short: "Delete a variable by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteVariable,
}

func init() {
	rootcmd.GetVariablesCmd().AddCommand(DeleteCmd)
}

func deleteVariable(cmd *cobra.Command, args []string) error {
	id := args[0]
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	client := n8n.NewClient(instanceURL, apiKey)
	if err := client.DeleteVariable(id); err != nil {
		return fmt.Errorf("failed to delete variable: %w", err)
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "Variable %q deleted successfully\n", id)
	return err
}
