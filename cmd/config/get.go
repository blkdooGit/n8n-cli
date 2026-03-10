package config

import (
	"fmt"
	"os"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var GetCmd = &cobra.Command{
	Use:   "get [KEY]",
	Short: "Get configuration values",
	Args:  cobra.MaximumNArgs(1),
	RunE:  getConfig,
}

func init() {
	rootcmd.GetConfigCmd().AddCommand(GetCmd)
	GetCmd.Flags().StringP("profile", "p", "", "Profile to read from")
}

func getConfig(cmd *cobra.Command, args []string) error {
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

	if len(args) == 0 && profile == "" {
		out, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(out))
		return err
	}

	if profile == "" {
		if dp, ok := config["default_profile"].(string); ok {
			profile = dp
		} else {
			profile = "default"
		}
	}

	profiles, _ := config["profiles"].(map[string]interface{})
	profileData, _ := profiles[profile].(map[string]interface{})

	if len(args) == 0 {
		out, err := yaml.Marshal(map[string]interface{}{profile: profileData})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(cmd.OutOrStdout(), string(out))
		return err
	}

	key := args[0]
	val, ok := profileData[key]
	if !ok {
		return fmt.Errorf("key %q not found in profile %q", key, profile)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%v\n", val)
	return err
}
