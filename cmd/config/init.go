package config

import (
	"fmt"
	"os"
	"path/filepath"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize n8n-cli configuration file",
	RunE:  initConfig,
}

func init() {
	rootcmd.GetConfigCmd().AddCommand(InitCmd)
	InitCmd.Flags().Bool("force", false, "Overwrite existing configuration")
}

const defaultConfig = `default_profile: default
profiles:
  default:
    instance_url: http://localhost:5678
    api_key: ""
    debug: false
`

func initConfig(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".n8n")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("configuration file already exists at %s (use --force to overwrite)", configPath)
	}

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Configuration initialized at %s\n", configPath)
	return err
}
