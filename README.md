<h1 align="center">N8N Command Line Interface (CLI)</h1>

<p align="center">
  <a href="https://github.com/edenreich/n8n-cli/actions/workflows/ci.yml">
    <img src="https://github.com/edenreich/n8n-cli/actions/workflows/ci.yml/badge.svg" alt="CI Status">
  </a>
  <a href="https://github.com/edenreich/n8n-cli/releases">
    <img src="https://img.shields.io/github/v/release/edenreich/n8n-cli" alt="Latest Release">
  </a>
  <a href="https://github.com/edenreich/n8n-cli/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/edenreich/n8n-cli" alt="License">
  </a>
  <a href="https://goreportcard.com/report/github.com/edenreich/n8n-cli">
    <img src="https://goreportcard.com/badge/github.com/edenreich/n8n-cli" alt="Go Report Card">
  </a>
  <a href="https://pkg.go.dev/github.com/edenreich/n8n-cli">
    <img src="https://pkg.go.dev/badge/github.com/edenreich/n8n-cli.svg" alt="Go Reference">
  </a>
</p>

<p align="center">Command line interface for managing n8n instances.</p>

## Table of Contents

- [Installation](#installation)
  - [Quick Install](#quick-install-linux-macos-windows-with-wsl)
  - [Autocompletion](#autocompletion)
  - [Manual Installation with Go](#manual-installation-with-go)
- [Configuration](#configuration)
  - [Profiles (multi-instance)](#profiles-multi-instance)
- [Commands](#commands)
  - [Version](#version)
  - [Workflows](#workflows)
    - [List](#list)
    - [Validate](#validate)
    - [Refresh](#refresh)
    - [Sync](#sync)
    - [Activate](#activate)
    - [Deactivate](#deactivate)
    - [Executions Retry](#executions-retry)
  - [Variables](#variables)
  - [Projects](#projects)
  - [Credentials](#credentials)
  - [Audit](#audit)
  - [Config](#config)
- [Development](#development)
- [Examples](#examples)
  - [Contact Form Example](#contact-form-example)
  - [AI-Enhanced Contact Form Example](#ai-enhanced-contact-form-example)

## Installation

### Quick Install (Linux, macOS, Windows with WSL)

```bash
curl -sSLf https://raw.github.com/edenreich/n8n-cli/main/install.sh | sh
```

Or install a specific version:

```bash
curl -sSLf https://raw.github.com/edenreich/n8n-cli/main/install.sh | sh -s -- --version v0.1.0-rc.1
```

This script will automatically detect your operating system and architecture and install the appropriate binary.

### Autocompletion

To enable auto completion for `bash`, `zsh`, or `fish`, run the following command:

```bash
source <(n8n completion bash) # for bash
source <(n8n completion zsh)  # for zsh
source <(n8n completion fish) # for fish
```

If you need it permanently, add it to your shell's configuration file (e.g., `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`).

### Manual Installation with Go

```bash
go install github.com/edenreich/n8n-cli@latest
```

## Configuration

Create a `.env` file in your current directory. The CLI will automatically load environment variables from this file.

```
N8N_API_KEY=your_n8n_api_key
N8N_INSTANCE_URL=https://your-instance.n8n.cloud
```

You can generate an API key in the n8n UI under Settings > API.

Alternatively, you can set these environment variables directly in your shell:

```bash
export N8N_API_KEY=your_n8n_api_key
export N8N_INSTANCE_URL=https://your-instance.n8n.cloud
```

Note: Environment variables set directly in your shell will take precedence over those defined in the `.env` file.

**Important:** Never commit your `.env` file containing API credentials to version control systems like GitHub. Make sure to add `.env` to your `.gitignore` file to prevent accidental exposure of sensitive credentials.

### Profiles (multi-instance)

The CLI supports named profiles for managing multiple n8n instances:

```bash
# Initialize config file
n8n config init

# Add a profile
n8n config profiles add production --url https://n8n.mycompany.com --api-key YOUR_KEY

# Set default profile
n8n config profiles use production

# Use a specific profile for a single command
n8n --profile production workflows list
```

Profile configuration is stored at `~/.n8n/config.yaml`.

## Commands

### Version

Display the version information of the n8n CLI:

```bash
n8n --version
# Or use the explicit command
n8n version
```

### Workflows

Manage n8n workflows with various subcommands.

#### List

List workflows from an n8n instance:

```bash
n8n workflows list
```

Options:

- `--output, -o`: Output format (default: "table"). Supported formats: `table`, `json`, `yaml`
- `--limit, -l`: Maximum number of workflows to return (default: 100, max: 250)
- `--name`: Filter by workflow name (partial match)
- `--tags`: Filter by tags (comma-separated)
- `--active`: Show only active workflows
- `--inactive`: Show only inactive workflows

Examples:

```bash
# List all workflows
n8n workflows list

# Filter by name and active status
n8n workflows list --name "invoice" --active

# Filter by tags
n8n workflows list --tags "production,crm"

# Output as JSON
n8n workflows list --output json
```

#### Validate

Statically analyze local workflow JSON/YAML files before syncing:

```bash
n8n workflows validate [FILES...] --directory workflows/
```

Options:

- `--directory, -d`: Validate all workflow files in a directory
- `--strict`: Fail on warnings in addition to errors

Examples:

```bash
# Validate all files in a directory
n8n workflows validate --directory workflows/

# Validate specific files
n8n workflows validate workflow1.json workflow2.yaml

# Strict mode (warnings are treated as errors)
n8n workflows validate --directory workflows/ --strict
```

#### Refresh

Refresh local workflow files with the current state from an n8n instance:

```bash
n8n workflows refresh --directory workflows/
```

The refresh command is an essential step before syncing to ensure you don't accidentally delete or overwrite workflows on the remote n8n instance. It pulls the current state of the workflows from n8n and updates or creates the corresponding local files.

Options:

- `--directory, -d`: Directory to store the workflow files (required)
- `--dry-run`: Show what would be updated without making changes
- `--overwrite`: Overwrite existing files even if they have a different name
- `--output, -o`: Output format for new workflow files (json or yaml)
- `--no-truncate`: Include all fields in output files, including null and optional fields (default: false)
- `--all`: Refresh all workflows from n8n instance, not just those in the directory.

Examples:

```bash
# Refresh only existing workflows in the directory
n8n workflows refresh --directory workflows/

# Refresh all workflows from n8n instance (including new ones)
n8n workflows refresh --directory workflows/ --all

# Preview what would be refreshed without making changes
n8n workflows refresh --directory workflows/ --dry-run

# Refresh workflows and save them as YAML files
n8n workflows refresh --directory workflows/ --output yaml

# Refresh workflows without minimizing the JSON/YAML output
n8n workflows refresh --directory workflows/ --no-truncate
```

#### Sync

Synchronize JSON workflows from a local directory to an n8n instance:

```bash
n8n workflows sync --directory workflows/
```

Options:

- `--directory, -d`: Directory containing workflow JSON/YAML files (required)
- `--dry-run`: Show what would be done without making changes
- `--prune`: Remove workflows from the n8n instance that are not present in the local directory
- `--refresh`: Refresh the local state with the remote state after sync (default: true)
- `--output, -o`: Output format for refreshed workflow files (json or yaml). If not specified, uses the existing file extension in the directory
- `--all`: Refresh all workflows from n8n instance when refreshing, not just those in the directory

How the sync command handles workflow IDs:

1. If a workflow file contains an ID:
   - If that ID exists on the n8n instance, the workflow will be updated
   - If that ID doesn't exist on the n8n instance, a new workflow will be created (n8n API doesn't allow specifying IDs when creating workflows)
2. If a workflow file doesn't have an ID, a new workflow will be created with a server-generated ID

This ensures that workflows maintain their IDs across different environments and prevents duplication.

Example:

```bash
# Sync workflows to the n8n instance
n8n workflows sync --directory workflows/

# Test without making changes
n8n workflows sync --directory workflows/ --dry-run

# Sync workflows and remove any remote workflows not in the local directory
n8n workflows sync --directory workflows/ --prune

# Sync workflows and refresh as JSON (overrides existing format)
n8n workflows sync --directory workflows/ --output json

# Sync workflows and refresh all workflows from n8n instance (including ones not in local directory)
n8n workflows sync --directory workflows/ --all

# Sync workflows without refreshing the local state afterward
n8n workflows sync --directory workflows/ --refresh=false
```

#### Activate

Activate a specific workflow by ID:

```bash
n8n workflows activate WORKFLOW_ID
```

#### Deactivate

Deactivate a specific workflow by ID:

```bash
n8n workflows deactivate WORKFLOW_ID
```

#### Executions Retry

Retry a failed execution:

```bash
n8n workflows executions retry EXECUTION_ID
```

Options:

- `--load-workflow`: Load the latest workflow version before retrying

### Variables

Manage environment variables in your n8n instance:

```bash
# List all variables
n8n variables list

# Set or update a variable
n8n variables set MY_KEY my_value

# Set with type
n8n variables set MY_NUMBER 42 --type number

# Delete a variable
n8n variables delete VARIABLE_ID

# Export to file (auto-detects format from extension)
n8n variables export --file vars.json
n8n variables export --file vars.yaml
n8n variables export --file .env

# Import from file
n8n variables import --file vars.json
n8n variables import --file .env
```

### Projects

Manage n8n projects:

```bash
# List all projects
n8n projects list

# Create a new project
n8n projects create "My Project"
```

### Credentials

Manage n8n credentials:

```bash
# Get the schema for a credential type
n8n credentials schema hubspotApi
n8n credentials schema slackApi

# Create a credential
n8n credentials create --name "My HubSpot" --type hubspotApi --data '{"apiKey":"your_key"}'

# Delete a credential
n8n credentials delete CREDENTIAL_ID
```

> **Note:** The n8n public API v1 does not expose a list credentials endpoint for security reasons. Use the n8n UI to view existing credentials.

### Audit

Generate a security audit report for your n8n instance:

```bash
# Generate audit report (table format)
n8n audit generate

# Output as JSON
n8n audit generate --output json

# Audit specific categories
n8n audit generate --categories credentials,nodes

# Configure abandoned workflow threshold
n8n audit generate --days-abandoned-workflow 30
```

### Config

Manage CLI configuration:

```bash
# Initialize config file at ~/.n8n/config.yaml
n8n config init

# Get current configuration
n8n config get
n8n config get instance_url

# Set a configuration value
n8n config set instance_url https://n8n.mycompany.com
n8n config set api_key YOUR_API_KEY

# Manage profiles
n8n config profiles list
n8n config profiles add staging --url https://staging.n8n.com --api-key KEY
n8n config profiles use staging
```

## Development

### Available Tasks

The project uses [Taskfile](https://taskfile.dev) for automating common development operations:

```bash
# Run unit tests
task test-unit

# Run integration tests
task test-integration

# Run all tests
task test-all

# Run linting
task lint

# Build the CLI
task build

# Run the CLI during development (args are passed to the CLI)
task cli -- workflows list
```

## Examples

The project includes practical examples to help you understand how to use the n8n-cli in real-world scenarios:

### Contact Form Example

A basic example that demonstrates how to set up a contact form workflow in n8n and synchronize it using the n8n-cli:

- HTML contact form
- n8n workflow for processing form submissions
- GitHub Actions workflow for automated synchronization

[View Contact Form Example](examples/contact-form/README.md)

### AI-Enhanced Contact Form Example

An advanced example that builds upon the basic contact form by adding AI capabilities:

- AI-powered message processing (summarization, sentiment analysis, categorization)
- Response suggestions generated by AI

[View AI-Enhanced Contact Form Example](examples/contact-form-ai/README.md)

These examples include complete workflow definitions, HTML templates, and detailed setup instructions.

## Contributing

We welcome contributions! Please see the [CONTRIBUTING.md](CONTRIBUTING.md) guide for details on how to set up the development environment, project structure, testing, and the pull request process.
