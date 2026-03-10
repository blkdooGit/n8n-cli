package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListProjectsCommand(t *testing.T) {
	fakeClient := &clientfakes.FakeClientInterface{}
	var stdout, stderr bytes.Buffer

	setupTestCommand := func() *cobra.Command {
		cmd := &cobra.Command{
			Use: "list",
			RunE: func(cmd *cobra.Command, args []string) error {
				outputFormat, _ := cmd.Flags().GetString("output")

				limit := 250
				var cursor string
				var allProjects []n8n.Project
				for {
					list, err := fakeClient.GetProjects(&limit, cursor)
					if err != nil {
						return err
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
					_, err = cmd.OutOrStdout().Write(append(data, '\n'))
					return err
				default:
					cmd.Println("ID\tNAME\tTYPE")
					for _, p := range allProjects {
						id := ""
						if p.Id != nil {
							id = *p.Id
						}
						pType := ""
						if p.Type != nil {
							pType = *p.Type
						}
						cmd.Printf("%s\t%s\t%s\n", id, p.Name, pType)
					}
					return nil
				}
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	createSampleProjectList := func(count int) *n8n.ProjectList {
		projects := make([]n8n.Project, count)
		for i := 0; i < count; i++ {
			id := "proj-" + string(rune('1'+i))
			pType := "personal"
			projects[i] = n8n.Project{
				Id:   &id,
				Name: "Project " + string(rune('A'+i)),
				Type: &pType,
			}
		}
		return &n8n.ProjectList{Data: &projects}
	}

	t.Run("successfully lists projects in table format", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetProjectsReturns(createSampleProjectList(3), nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Project A")
		assert.Contains(t, stdout.String(), "Project B")
		assert.Contains(t, stdout.String(), "Project C")
		assert.Contains(t, stdout.String(), "proj-")
	})

	t.Run("shows empty output when no projects exist", func(t *testing.T) {
		cmd := setupTestCommand()
		emptyData := []n8n.Project{}
		fakeClient.GetProjectsReturns(&n8n.ProjectList{Data: &emptyData}, nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		// Header row is printed but no project entries
		assert.Contains(t, stdout.String(), "ID")
		assert.NotContains(t, stdout.String(), "proj-")
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetProjectsReturns(nil, errors.New("API error"))

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("outputs JSON when --output json is specified", func(t *testing.T) {
		cmd := setupTestCommand()
		err := cmd.Flags().Set("output", "json")
		assert.NoError(t, err)

		fakeClient.GetProjectsReturns(createSampleProjectList(2), nil)

		err = cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), `"name"`)
		assert.Contains(t, stdout.String(), "Project A")
	})
}
