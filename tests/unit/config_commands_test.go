package unit

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	configcmd "github.com/edenreich/n8n-cli/cmd/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// setTempHome redirects HOME to a temp directory for the duration of the test
// so that config commands write/read from an isolated location.
func setTempHome(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "n8n-config-test-*")
	assert.NoError(t, err)

	origHome := os.Getenv("HOME")
	err = os.Setenv("HOME", tmpDir)
	assert.NoError(t, err)

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

// writeTestConfig writes a minimal config file into the given home directory
// and returns the path to the config file.
func writeTestConfig(t *testing.T, homeDir string) string {
	t.Helper()
	configDir := filepath.Join(homeDir, ".n8n")
	err := os.MkdirAll(configDir, 0700)
	assert.NoError(t, err)

	content := `default_profile: default
profiles:
  default:
    instance_url: http://localhost:5678
    api_key: ""
    debug: false
`
	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(content), 0600)
	assert.NoError(t, err)
	return configPath
}

// newTestCmd wraps a real RunE into a fresh cobra.Command with the provided flags
// redirected to controlled stdout/stderr buffers.
func newTestCmd(stdout, stderr *bytes.Buffer, runE func(*cobra.Command, []string) error) *cobra.Command {
	cmd := &cobra.Command{RunE: runE}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	stdout.Reset()
	stderr.Reset()
	return cmd
}

// ---- config init tests ----

func TestConfigInitCommand(t *testing.T) {
	t.Run("creates config file in HOME/.n8n", func(t *testing.T) {
		homeDir := setTempHome(t)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.InitCmd.RunE)
		cmd.Flags().Bool("force", false, "")

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "Configuration initialized at")

		configPath := filepath.Join(homeDir, ".n8n", "config.yaml")
		data, err := os.ReadFile(configPath)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "default_profile: default")
		assert.Contains(t, string(data), "instance_url: http://localhost:5678")
	})

	t.Run("does not overwrite existing config without --force", func(t *testing.T) {
		homeDir := setTempHome(t)
		configPath := writeTestConfig(t, homeDir)

		customContent := "custom: preserved\n"
		err := os.WriteFile(configPath, []byte(customContent), 0600)
		assert.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.InitCmd.RunE)
		cmd.Flags().Bool("force", false, "")

		err = cmd.Execute()

		// Should return an error when file exists and --force is not set
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		data, readErr := os.ReadFile(configPath)
		assert.NoError(t, readErr)
		assert.Equal(t, customContent, string(data), "file should not be overwritten")
	})

	t.Run("overwrites existing config with --force", func(t *testing.T) {
		homeDir := setTempHome(t)
		configPath := writeTestConfig(t, homeDir)

		err := os.WriteFile(configPath, []byte("custom: preserved\n"), 0600)
		assert.NoError(t, err)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.InitCmd.RunE)
		cmd.Flags().Bool("force", false, "")
		err = cmd.Flags().Set("force", "true")
		assert.NoError(t, err)

		err = cmd.Execute()

		assert.NoError(t, err)
		data, readErr := os.ReadFile(configPath)
		assert.NoError(t, readErr)
		assert.Contains(t, string(data), "default_profile: default")
	})
}

// ---- config set tests ----

func TestConfigSetCommand(t *testing.T) {
	t.Run("sets a key in the default profile", func(t *testing.T) {
		homeDir := setTempHome(t)
		configPath := writeTestConfig(t, homeDir)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.SetCmd.RunE)
		cmd.Flags().StringP("profile", "p", "", "")

		cmd.SetArgs([]string{"api_key", "my-new-key"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "default.api_key = my-new-key")

		data, readErr := os.ReadFile(configPath)
		assert.NoError(t, readErr)
		assert.Contains(t, string(data), "my-new-key")
	})

	t.Run("sets a key in a specific profile with --profile", func(t *testing.T) {
		homeDir := setTempHome(t)
		configPath := writeTestConfig(t, homeDir)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.SetCmd.RunE)
		cmd.Flags().StringP("profile", "p", "", "")
		err := cmd.Flags().Set("profile", "staging")
		assert.NoError(t, err)

		cmd.SetArgs([]string{"instance_url", "http://staging:5678"})
		err = cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "staging.instance_url = http://staging:5678")

		data, readErr := os.ReadFile(configPath)
		assert.NoError(t, readErr)
		assert.Contains(t, string(data), "http://staging:5678")
	})

	t.Run("returns error when config file does not exist", func(t *testing.T) {
		setTempHome(t) // empty home, no config file

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.SetCmd.RunE)
		cmd.Flags().StringP("profile", "p", "", "")

		cmd.SetArgs([]string{"api_key", "value"})
		err := cmd.Execute()

		assert.Error(t, err)
	})
}

// ---- config get tests ----

func TestConfigGetCommand(t *testing.T) {
	t.Run("gets a specific key from the default profile", func(t *testing.T) {
		homeDir := setTempHome(t)
		writeTestConfig(t, homeDir)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.GetCmd.RunE)
		cmd.Flags().StringP("profile", "p", "", "")

		cmd.SetArgs([]string{"instance_url"})
		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "http://localhost:5678")
	})

	t.Run("prints entire config when no args given", func(t *testing.T) {
		homeDir := setTempHome(t)
		writeTestConfig(t, homeDir)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.GetCmd.RunE)
		cmd.Flags().StringP("profile", "p", "", "")

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "default_profile")
	})

	t.Run("returns error for non-existent key", func(t *testing.T) {
		homeDir := setTempHome(t)
		writeTestConfig(t, homeDir)

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.GetCmd.RunE)
		cmd.Flags().StringP("profile", "p", "", "")

		cmd.SetArgs([]string{"nonexistent_key"})
		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent_key")
	})

	t.Run("returns error when config file does not exist", func(t *testing.T) {
		setTempHome(t) // empty home, no config file

		var stdout, stderr bytes.Buffer
		cmd := newTestCmd(&stdout, &stderr, configcmd.GetCmd.RunE)
		cmd.Flags().StringP("profile", "p", "", "")

		err := cmd.Execute()

		assert.Error(t, err)
	})
}
