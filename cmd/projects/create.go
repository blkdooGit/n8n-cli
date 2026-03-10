package projects

import (
	"fmt"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CreateCmd = &cobra.Command{
	Use:   "create <NAME>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	RunE:  createProject,
}

func init() {
	rootcmd.GetProjectsCmd().AddCommand(CreateCmd)
}

func createProject(cmd *cobra.Command, args []string) error {
	name := args[0]
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	client := n8n.NewClient(instanceURL, apiKey)
	if err := client.CreateProject(name); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "Project %q created successfully\n", name)
	return err
}
