package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"text/tabwriter"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestGenerateAuditCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	setupTestCommand := func(fakeClient *clientfakes.FakeClientInterface) *cobra.Command {
		cmd := &cobra.Command{
			Use:  "generate",
			RunE: func(cmd *cobra.Command, args []string) error {
				outputFormat, _ := cmd.Flags().GetString("output")
				categories, _ := cmd.Flags().GetStringSlice("categories")
				daysAbandoned, _ := cmd.Flags().GetInt("days-abandoned-workflow")

				var options *n8n.PostAuditJSONBody
				if len(categories) > 0 || daysAbandoned != 90 {
					cats := make([]n8n.PostAuditJSONBodyAdditionalOptionsCategories, 0, len(categories))
					for _, c := range categories {
						cats = append(cats, n8n.PostAuditJSONBodyAdditionalOptionsCategories(c))
					}
					options = &n8n.PostAuditJSONBody{
						AdditionalOptions: &struct {
							Categories *[]n8n.PostAuditJSONBodyAdditionalOptionsCategories `json:"categories,omitempty"`
							DaysAbandonedWorkflow *int                                     `json:"daysAbandonedWorkflow,omitempty"`
						}{},
					}
					if len(cats) > 0 {
						options.AdditionalOptions.Categories = &cats
					}
					options.AdditionalOptions.DaysAbandonedWorkflow = &daysAbandoned
				}

				report, err := fakeClient.GenerateAudit(options)
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
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")
		cmd.Flags().StringSlice("categories", nil, "Audit categories to check")
		cmd.Flags().Int("days-abandoned-workflow", 90, "Days after which a workflow is considered abandoned")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	buildSampleAudit := func() *n8n.Audit {
		credReport := map[string]interface{}{
			"risk1": "exposed credential",
			"risk2": "weak password",
		}
		nodesReport := map[string]interface{}{
			"risk1": "outdated node",
		}
		return &n8n.Audit{
			CredentialsRiskReport: &credReport,
			NodesRiskReport:       &nodesReport,
		}
	}

	t.Run("successfully generates audit report (table format)", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GenerateAuditReturns(buildSampleAudit(), nil)

		err := cmd.Execute()
		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "SECTION")
		assert.Contains(t, stdout.String(), "RISK COUNT")
		assert.Contains(t, stdout.String(), "Credentials Risk Report")
		assert.Contains(t, stdout.String(), "Nodes Risk Report")
		assert.Equal(t, 1, fakeClient.GenerateAuditCallCount())
	})

	t.Run("successfully generates audit report (JSON format)", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GenerateAuditReturns(buildSampleAudit(), nil)

		err := cmd.Flags().Set("output", "json")
		assert.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)

		var result n8n.Audit
		err = json.Unmarshal([]byte(stdout.String()), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result.CredentialsRiskReport)
		assert.NotNil(t, result.NodesRiskReport)
		assert.Equal(t, 1, fakeClient.GenerateAuditCallCount())
	})

	t.Run("returns error when API fails", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GenerateAuditReturns(nil, errors.New("audit service unavailable"))

		err := cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audit service unavailable")
		assert.Equal(t, 1, fakeClient.GenerateAuditCallCount())
	})

	t.Run("passes nil options when using defaults", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GenerateAuditReturns(buildSampleAudit(), nil)

		// Execute with defaults (no categories, days-abandoned-workflow=90)
		err := cmd.Execute()
		assert.NoError(t, err)

		passedOptions := fakeClient.GenerateAuditArgsForCall(0)
		assert.Nil(t, passedOptions, "options should be nil when using defaults")
	})

	t.Run("passes categories options when specified", func(t *testing.T) {
		fakeClient := &clientfakes.FakeClientInterface{}
		cmd := setupTestCommand(fakeClient)

		fakeClient.GenerateAuditReturns(buildSampleAudit(), nil)

		err := cmd.Flags().Set("categories", "credentials,nodes")
		assert.NoError(t, err)

		err = cmd.Execute()
		assert.NoError(t, err)

		passedOptions := fakeClient.GenerateAuditArgsForCall(0)
		assert.NotNil(t, passedOptions)
		assert.NotNil(t, passedOptions.AdditionalOptions)
		assert.NotNil(t, passedOptions.AdditionalOptions.Categories)
		assert.Len(t, *passedOptions.AdditionalOptions.Categories, 2)
	})
}
