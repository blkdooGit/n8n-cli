package variables

import (
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

var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export variables to a file",
	RunE:  exportVariables,
}

func init() {
	rootcmd.GetVariablesCmd().AddCommand(ExportCmd)
	ExportCmd.Flags().StringP("file", "f", "", "Output file path (required)")
	ExportCmd.Flags().StringP("format", "", "", "Output format: json, yaml, dotenv (auto-detected from file extension)")
	ExportCmd.MarkFlagRequired("file")
}

func exportVariables(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	format, _ := cmd.Flags().GetString("format")
	instanceURL := viper.GetString("instance_url")
	apiKey := viper.GetString("api_key")

	if format == "" {
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".yaml", ".yml":
			format = "yaml"
		case ".env":
			format = "dotenv"
		default:
			format = "json"
		}
	}

	client := n8n.NewClient(instanceURL, apiKey)
	var allVariables []n8n.Variable
	limit := 250
	var cursor string
	for {
		varList, err := client.GetVariables(&limit, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch variables: %w", err)
		}
		if varList.Data != nil {
			allVariables = append(allVariables, *varList.Data...)
		}
		if varList.NextCursor == nil || *varList.NextCursor == "" {
			break
		}
		cursor = *varList.NextCursor
	}

	var content []byte
	var err error
	switch format {
	case "yaml":
		content, err = yaml.Marshal(allVariables)
	case "dotenv":
		var sb strings.Builder
		for _, v := range allVariables {
			sb.WriteString(fmt.Sprintf("%s=%s\n", v.Key, v.Value))
		}
		content = []byte(sb.String())
	default:
		content, err = json.MarshalIndent(allVariables, "", "  ")
	}
	if err != nil {
		return fmt.Errorf("failed to serialize variables: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Exported %d variables to %s\n", len(allVariables), filePath)
	return err
}
