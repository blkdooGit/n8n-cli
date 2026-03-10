package credentials

import (
	"encoding/json"
	"fmt"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SchemaCmd = &cobra.Command{
	Use:   "schema <TYPE>",
	Short: "Get the schema for a credential type",
	Args:  cobra.ExactArgs(1),
	RunE:  getSchema,
}

func init() {
	rootcmd.GetCredentialsCmd().AddCommand(SchemaCmd)
}

func getSchema(cmd *cobra.Command, args []string) error {
	credType := args[0]
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	client := n8n.NewClient(instanceURL, apiKey)
	schema, err := client.GetCredentialSchema(credType)
	if err != nil {
		return fmt.Errorf("failed to get schema for %q: %w", credType, err)
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return err
}
