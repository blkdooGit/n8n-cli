package credentials

import (
	"fmt"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var DeleteCredCmd = &cobra.Command{
	Use:   "delete <ID>",
	Short: "Delete a credential by ID",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteCredential,
}

func init() {
	rootcmd.GetCredentialsCmd().AddCommand(DeleteCredCmd)
}

func deleteCredential(cmd *cobra.Command, args []string) error {
	id := args[0]
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	client := n8n.NewClient(instanceURL, apiKey)
	if err := client.DeleteCredential(id); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "Credential %q deleted successfully\n", id)
	return err
}
