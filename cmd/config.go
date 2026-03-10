package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage n8n-cli configuration",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func GetConfigCmd() *cobra.Command { return configCmd }
