package credentials

import (
	"encoding/json"
	"fmt"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new credential",
	RunE:  createCredential,
}

func init() {
	rootcmd.GetCredentialsCmd().AddCommand(CreateCmd)
	CreateCmd.Flags().StringP("name", "n", "", "Credential name (required)")
	CreateCmd.Flags().StringP("type", "t", "", "Credential type (required)")
	CreateCmd.Flags().StringP("data", "d", "", "Credential data as JSON string (required)")
	CreateCmd.MarkFlagRequired("name")
	CreateCmd.MarkFlagRequired("type")
	CreateCmd.MarkFlagRequired("data")
}

func createCredential(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	credType, _ := cmd.Flags().GetString("type")
	dataStr, _ := cmd.Flags().GetString("data")

	var credData map[string]interface{}
	if err := json.Unmarshal([]byte(dataStr), &credData); err != nil {
		return fmt.Errorf("invalid JSON in --data: %w", err)
	}

	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")
	client := n8n.NewClient(instanceURL, apiKey)

	cred := &n8n.Credential{
		Name: name,
		Type: credType,
		Data: &credData,
	}

	resp, err := client.CreateCredential(cred)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	id := ""
	if resp != nil && resp.Id != nil {
		id = *resp.Id
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Credential %q created with ID: %s\n", name, id)
	return err
}
