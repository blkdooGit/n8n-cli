package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var SetCmd = &cobra.Command{
	Use:   "set <KEY> <VALUE>",
	Short: "Set a configuration value",
	Long:  "Set a configuration value. KEY can be: instance_url, api_key, debug. Use --profile to target a specific profile.",
	Args:  cobra.ExactArgs(2),
	RunE:  setConfig,
}

func init() {
	rootcmd.GetConfigCmd().AddCommand(SetCmd)
	SetCmd.Flags().StringP("profile", "p", "", "Target profile (default: default_profile from config)")
}

func setConfig(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]
	profile, _ := cmd.Flags().GetString("profile")

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("config file not found. Run 'n8n config init' first: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if profile == "" {
		if dp, ok := config["default_profile"].(string); ok {
			profile = dp
		} else {
			profile = "default"
		}
	}

	profiles, _ := config["profiles"].(map[string]interface{})
	if profiles == nil {
		profiles = make(map[string]interface{})
		config["profiles"] = profiles
	}

	profileData, _ := profiles[profile].(map[string]interface{})
	if profileData == nil {
		profileData = make(map[string]interface{})
		profiles[profile] = profileData
	}

	profileData[key] = parseValue(value)

	out, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configPath, out, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Set %s.%s = %s\n", profile, key, value)
	return err
}

func parseValue(s string) interface{} {
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	}
	return s
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".n8n", "config.yaml"), nil
}
