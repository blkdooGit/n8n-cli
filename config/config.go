// Package config provides configuration functionality for the n8n-cli application
package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Initialize reads in config file and ENV variables if set
func Initialize() {
	LoadEnvFile()

	v := viper.GetViper()
	v.SetEnvPrefix("N8N")
	v.AutomaticEnv()

	BindEnvSafely(v, "api_key", "N8N_API_KEY")
	BindEnvSafely(v, "instance_url", "N8N_INSTANCE_URL")

	v.SetDefault("instance_url", "http://localhost:5678")
	v.SetDefault("api_key", "")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("$HOME/.n8n")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: Config file error: %v\n", err)
		}
	}
}

// BindEnvSafely binds an environment variable and logs errors without crashing
func BindEnvSafely(v *viper.Viper, key, envVar string) {
	if err := v.BindEnv(key, envVar); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding environment variable %s: %v\n", envVar, err)
	}
}

// LoadProfile loads a named profile from the config file into viper
func LoadProfile(profileName string) error {
	profiles := viper.GetStringMap("profiles")
	profileData, ok := profiles[profileName]
	if !ok {
		return fmt.Errorf("profile %q not found in config", profileName)
	}

	profile, ok := profileData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid profile format for %q", profileName)
	}

	// Only set values that aren't already set by flags/env
	for k, v := range profile {
		if !viper.IsSet(k) {
			viper.Set(k, v)
		}
	}
	return nil
}
