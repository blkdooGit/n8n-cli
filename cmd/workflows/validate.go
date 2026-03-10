package workflows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	rootcmd "github.com/edenreich/n8n-cli/cmd"
	"github.com/edenreich/n8n-cli/n8n"
	"github.com/spf13/cobra"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate [FILES...]",
	Short: "Validate local workflow JSON/YAML files",
	Long:  `Statically analyzes workflow files to detect structural issues before syncing.`,
	RunE:  validateWorkflows,
}

func init() {
	rootcmd.GetWorkflowsCmd().AddCommand(ValidateCmd)
	ValidateCmd.Flags().StringP("directory", "d", "", "Validate all workflow files in a directory")
	ValidateCmd.Flags().Bool("strict", false, "Fail on warnings in addition to errors")
}

// ValidationResult holds the result of validating a single file
type ValidationResult struct {
	File   string
	Status string
	Issues []string
}

func validateWorkflows(cmd *cobra.Command, args []string) error {
	directory, _ := cmd.Flags().GetString("directory")
	strict, _ := cmd.Flags().GetBool("strict")

	var files []string
	if directory != "" {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".json" || ext == ".yaml" || ext == ".yml" {
				files = append(files, filepath.Join(directory, e.Name()))
			}
		}
	}
	files = append(files, args...)

	if len(files) == 0 {
		return fmt.Errorf("no files to validate. Provide file paths or use --directory")
	}

	decoder := n8n.NewWorkflowDecoder()
	var results []ValidationResult
	hasErrors := false
	hasWarnings := false

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			results = append(results, ValidationResult{File: f, Status: "ERROR", Issues: []string{fmt.Sprintf("cannot read file: %v", err)}})
			hasErrors = true
			continue
		}

		workflow, err := decoder.DecodeFromBytes(data)
		if err != nil {
			results = append(results, ValidationResult{File: f, Status: "ERROR", Issues: []string{fmt.Sprintf("parse error: %v", err)}})
			hasErrors = true
			continue
		}

		var issues []string
		var warnings []string

		if workflow.Name == "" {
			issues = append(issues, "workflow name is empty")
		}

		if len(workflow.Nodes) == 0 {
			issues = append(issues, "workflow has no nodes")
		}

		nodeNames := make(map[string]int)
		for i, node := range workflow.Nodes {
			if node.Name == nil || *node.Name == "" {
				issues = append(issues, fmt.Sprintf("node[%d] has no name", i))
			} else {
				nodeNames[*node.Name]++
			}
			if node.Type == nil || *node.Type == "" {
				issues = append(issues, fmt.Sprintf("node[%d] has no type", i))
			}
		}

		for name, count := range nodeNames {
			if count > 1 {
				issues = append(issues, fmt.Sprintf("duplicate node name: %q (appears %d times)", name, count))
			}
		}

		if workflow.Connections != nil {
			for connName := range workflow.Connections {
				if _, exists := nodeNames[connName]; !exists {
					warnings = append(warnings, fmt.Sprintf("connection references unknown node: %q", connName))
				}
			}
		}

		allIssues := issues
		status := "OK"
		if len(issues) > 0 {
			status = "ERROR"
			hasErrors = true
		} else if len(warnings) > 0 {
			allIssues = warnings
			status = "WARNING"
			hasWarnings = true
		}

		results = append(results, ValidationResult{File: f, Status: status, Issues: allIssues})
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "FILE\tSTATUS\tISSUES")
	for _, r := range results {
		issues := strings.Join(r.Issues, "; ")
		if issues == "" {
			issues = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.File, r.Status, issues)
	}
	w.Flush()

	if hasErrors || (strict && hasWarnings) {
		return fmt.Errorf("validation failed")
	}
	return nil
}
