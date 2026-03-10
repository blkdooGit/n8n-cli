package variables

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import variables from a file",
	RunE:  importVariables,
}

func init() {
	rootcmd.GetVariablesCmd().AddCommand(ImportCmd)
	ImportCmd.Flags().StringP("file", "f", "", "Input file path (required)")
	ImportCmd.MarkFlagRequired("file")
}

func importVariables(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var variables []n8n.Variable
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &variables); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	case ".env":
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			variables = append(variables, n8n.Variable{
				Key:   strings.TrimSpace(parts[0]),
				Value: strings.TrimSpace(parts[1]),
			})
		}
	default:
		if err := json.Unmarshal(data, &variables); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	client := n8n.NewClient(instanceURL, apiKey)

	// Fetch existing variables for upsert
	limit := 250
	existingMap := make(map[string]*n8n.Variable)
	var cursor string
	for {
		varList, err := client.GetVariables(&limit, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch existing variables: %w", err)
		}
		if varList.Data != nil {
			for i := range *varList.Data {
				v := &(*varList.Data)[i]
				existingMap[v.Key] = v
			}
		}
		if varList.NextCursor == nil || *varList.NextCursor == "" {
			break
		}
		cursor = *varList.NextCursor
	}

	created, updated := 0, 0
	for i := range variables {
		v := &variables[i]
		if existing, found := existingMap[v.Key]; found {
			if err := client.UpdateVariable(*existing.Id, v); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to update %q: %v\n", v.Key, err)
				continue
			}
			updated++
		} else {
			if err := client.CreateVariable(v); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to create %q: %v\n", v.Key, err)
				continue
			}
			created++
		}
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Imported variables: %d created, %d updated\n", created, updated)
	return err
}
