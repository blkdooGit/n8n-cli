package projects

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
	Short: "List all projects",
	RunE:  listProjects,
}

func init() {
	rootcmd.GetProjectsCmd().AddCommand(ListCmd)
	ListCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
}

func listProjects(cmd *cobra.Command, args []string) error {
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")
	outputFormat, _ := cmd.Flags().GetString("output")

	client := n8n.NewClient(instanceURL, apiKey)
	limit := 250
	var cursor string
	var allProjects []n8n.Project
	for {
		list, err := client.GetProjects(&limit, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch projects: %w", err)
		}
		if list.Data != nil {
			allProjects = append(allProjects, *list.Data...)
		}
		if list.NextCursor == nil || *list.NextCursor == "" {
			break
		}
		cursor = *list.NextCursor
	}

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(allProjects, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	case "yaml":
		data, err := yaml.Marshal(allProjects)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(data))
		return err
	default:
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTYPE")
		for _, p := range allProjects {
			id, name, pType := "", p.Name, ""
			if p.Id != nil {
				id = *p.Id
			}
			if p.Type != nil {
				pType = *p.Type
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", id, name, pType)
		}
		return w.Flush()
	}
}
