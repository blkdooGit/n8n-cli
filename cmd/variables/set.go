package variables

import (
	"fmt"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var SetCmd = &cobra.Command{
	Use:   "set <KEY> <VALUE>",
	Short: "Set a variable value (creates or updates)",
	Args:  cobra.ExactArgs(2),
	RunE:  setVariable,
}

func init() {
	rootcmd.GetVariablesCmd().AddCommand(SetCmd)
	SetCmd.Flags().StringP("type", "t", "string", "Variable type (string, number, boolean, json)")
}

func setVariable(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]
	varType, _ := cmd.Flags().GetString("type")
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	client := n8n.NewClient(instanceURL, apiKey)

	// Find existing variable by key
	limit := 250
	var cursor string
	var existingID *string
	for {
		varList, err := client.GetVariables(&limit, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch variables: %w", err)
		}
		if varList.Data != nil {
			for _, v := range *varList.Data {
				if v.Key == key {
					existingID = v.Id
					break
				}
			}
		}
		if existingID != nil || varList.NextCursor == nil || *varList.NextCursor == "" {
			break
		}
		cursor = *varList.NextCursor
	}

	variable := &n8n.Variable{
		Key:   key,
		Value: value,
		Type:  &varType,
	}

	if existingID != nil {
		if err := client.UpdateVariable(*existingID, variable); err != nil {
			return fmt.Errorf("failed to update variable: %w", err)
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "Variable %q updated successfully\n", key)
		return err
	}

	if err := client.CreateVariable(variable); err != nil {
		return fmt.Errorf("failed to create variable: %w", err)
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "Variable %q created successfully\n", key)
	return err
}
