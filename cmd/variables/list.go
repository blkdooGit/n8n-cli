package variables

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all variables from n8n instance",
	RunE:  listVariables,
}

func init() {
	rootcmd.GetVariablesCmd().AddCommand(ListCmd)
	ListCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
}

func listVariables(cmd *cobra.Command, args []string) error {
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")
	outputFormat, _ := cmd.Flags().GetString("output")

	client := n8n.NewClient(instanceURL, apiKey)

	var allVariables []n8n.Variable
	var cursor string
	limit := 100
	for {
		varList, err := client.GetVariables(&limit, cursor)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error fetching variables: %v\n", err)
			return err
		}
		if varList.Data != nil {
			allVariables = append(allVariables, *varList.Data...)
		}
		if varList.NextCursor == nil || *varList.NextCursor == "" {
			break
		}
		cursor = *varList.NextCursor
	}

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(allVariables, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	case "yaml":
		data, err := yaml.Marshal(allVariables)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(data))
		return err
	default:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "KEY\tVALUE\tTYPE\tID")
		for _, v := range allVariables {
			id := ""
			if v.Id != nil {
				id = *v.Id
			}
			varType := ""
			if v.Type != nil {
				varType = *v.Type
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", v.Key, v.Value, varType, id)
		}
		return w.Flush()
	}
}
