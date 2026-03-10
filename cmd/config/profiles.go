package config

import (
	"fmt"
	"os"
	"text/tabwriter"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var ProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "Manage configuration profiles",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration profiles",
	RunE:  listProfiles,
}

var profilesAddCmd = &cobra.Command{
	Use:   "add <NAME>",
	Short: "Add a new configuration profile",
	Args:  cobra.ExactArgs(1),
	RunE:  addProfile,
}

var profilesUseCmd = &cobra.Command{
	Use:   "use <NAME>",
	Short: "Set the default profile",
	Args:  cobra.ExactArgs(1),
	RunE:  useProfile,
}

func init() {
	rootcmd.GetConfigCmd().AddCommand(ProfilesCmd)
	ProfilesCmd.AddCommand(profilesListCmd, profilesAddCmd, profilesUseCmd)
	profilesAddCmd.Flags().StringP("url", "u", "http://localhost:5678", "n8n instance URL")
	profilesAddCmd.Flags().StringP("api-key", "k", "", "n8n API key")
}

func listProfiles(cmd *cobra.Command, args []string) error {
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
		return err
	}
	defaultProfile, _ := config["default_profile"].(string)
	profiles, _ := config["profiles"].(map[string]interface{})

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tURL\tDEFAULT")
	for name, pd := range profiles {
		profile, _ := pd.(map[string]interface{})
		url, _ := profile["instance_url"].(string)
		isDefault := ""
		if name == defaultProfile {
			isDefault = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, url, isDefault)
	}
	return w.Flush()
}

func addProfile(cmd *cobra.Command, args []string) error {
	name := args[0]
	url, _ := cmd.Flags().GetString("url")
	apiKey, _ := cmd.Flags().GetString("api-key")

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
		return err
	}
	profiles, _ := config["profiles"].(map[string]interface{})
	if profiles == nil {
		profiles = make(map[string]interface{})
		config["profiles"] = profiles
	}
	profiles[name] = map[string]interface{}{
		"instance_url": url,
		"api_key":      apiKey,
		"debug":        false,
	}

	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, out, 0600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Profile %q added\n", name)
	return err
}

func useProfile(cmd *cobra.Command, args []string) error {
	name := args[0]
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
		return err
	}
	profiles, _ := config["profiles"].(map[string]interface{})
	if _, ok := profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	config["default_profile"] = name
	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, out, 0600); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Default profile set to %q\n", name)
	return err
}
