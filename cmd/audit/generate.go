package audit

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

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a security audit report",
	RunE:  generateAudit,
}

func init() {
	rootcmd.GetAuditCmd().AddCommand(GenerateCmd)
	GenerateCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
	GenerateCmd.Flags().StringSlice("categories", nil, "Audit categories to check (credentials, database, filesystem, nodes, instance)")
	GenerateCmd.Flags().Int("days-abandoned-workflow", 90, "Days after which a workflow is considered abandoned")
}

func generateAudit(cmd *cobra.Command, args []string) error {
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")
	outputFormat, _ := cmd.Flags().GetString("output")
	categories, _ := cmd.Flags().GetStringSlice("categories")
	daysAbandoned, _ := cmd.Flags().GetInt("days-abandoned-workflow")

	client := n8n.NewClient(instanceURL, apiKey)

	// Build options from flags
	var options *n8n.PostAuditJSONBody
	if len(categories) > 0 || daysAbandoned != 90 {
		cats := make([]n8n.PostAuditJSONBodyAdditionalOptionsCategories, 0, len(categories))
		for _, c := range categories {
			cats = append(cats, n8n.PostAuditJSONBodyAdditionalOptionsCategories(c))
		}
		options = &n8n.PostAuditJSONBody{
			AdditionalOptions: &struct {
				Categories *[]n8n.PostAuditJSONBodyAdditionalOptionsCategories `json:"categories,omitempty"`

				// DaysAbandonedWorkflow Days for a workflow to be considered abandoned if not executed
				DaysAbandonedWorkflow *int `json:"daysAbandonedWorkflow,omitempty"`
			}{},
		}
		if len(cats) > 0 {
			options.AdditionalOptions.Categories = &cats
		}
		options.AdditionalOptions.DaysAbandonedWorkflow = &daysAbandoned
	}

	report, err := client.GenerateAudit(options)
	if err != nil {
		return fmt.Errorf("failed to generate audit: %w", err)
	}

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	case "yaml":
		data, err := yaml.Marshal(report)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(data))
		return err
	default:
		return printAuditTable(cmd, report)
	}
}

func printAuditTable(cmd *cobra.Command, report *n8n.Audit) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SECTION\tRISK COUNT")
	if report.CredentialsRiskReport != nil {
		fmt.Fprintf(w, "Credentials Risk Report\t%d\n", len(*report.CredentialsRiskReport))
	}
	if report.DatabaseRiskReport != nil {
		fmt.Fprintf(w, "Database Risk Report\t%d\n", len(*report.DatabaseRiskReport))
	}
	if report.FilesystemRiskReport != nil {
		fmt.Fprintf(w, "Filesystem Risk Report\t%d\n", len(*report.FilesystemRiskReport))
	}
	if report.InstanceRiskReport != nil {
		fmt.Fprintf(w, "Instance Risk Report\t%d\n", len(*report.InstanceRiskReport))
	}
	if report.NodesRiskReport != nil {
		fmt.Fprintf(w, "Nodes Risk Report\t%d\n", len(*report.NodesRiskReport))
	}
	return w.Flush()
}
