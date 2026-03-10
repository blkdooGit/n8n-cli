package cmd

import "github.com/spf13/cobra"

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Generate security audit report for n8n instance",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

func init() {
	rootCmd.AddCommand(auditCmd)
}

func GetAuditCmd() *cobra.Command { return auditCmd }
