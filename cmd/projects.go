package cmd

import "github.com/spf13/cobra"

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage n8n projects",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}

func GetProjectsCmd() *cobra.Command { return projectsCmd }
