package cmd

import "github.com/spf13/cobra"

var credentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Manage n8n credentials",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

func init() {
	rootCmd.AddCommand(credentialsCmd)
}

func GetCredentialsCmd() *cobra.Command { return credentialsCmd }
