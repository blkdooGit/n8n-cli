# N8N-CLI TODO List

This document outlines essential features for the n8n-cli tool based on the generated OpenAPI types. These features cover the core functionality needed to effectively manage n8n workflows between local development and n8n instances.

## Core Features

### Workflow Management

- [x] Import workflows from n8n instance to local files
- [x] Synchronize local workflow files to n8n instance
- [x] Activate/deactivate workflows
- [x] Delete workflows from n8n instance
- [x] Get execution history for workflows
- [x] Implement validate command to apply static analysis on workflow files - `n8n workflows validate`
- [x] List all workflows with filter capabilities (by name, tags, active status) - `--name`, `--tags`, `--active`, `--inactive` flags

### Tags Management

- [x] Add/remove tags to workflows
- [x] Create new tags

## Documentation

- [x] Generate command reference
- [x] Create examples for common workflows
- [x] Document best practices for workflow version control

### Credentials Management

- [x] Create credentials - `n8n credentials create --name <name> --type <type> --data <json>`
- [x] Delete credentials - `n8n credentials delete <id>`
- [x] Get credential schema - `n8n credentials schema <type>`
- [ ] List credentials from n8n instance - **Not possible via public API v1** (no GET /credentials endpoint)

### Workflow Execution

- [ ] Execute a workflow manually - **Not possible via public API v1** (no POST /workflows/:id/run endpoint)
- [x] Retrieve execution results - `n8n workflows executions <id>` with `--include-data`
- [x] Retry failed executions - `n8n workflows executions retry <id>`
- [x] Delete executions - via `DeleteExecution` API method

### Variables Management

- [x] List variables - `n8n variables list`
- [x] Export variables to local files - `n8n variables export --file <path>`
- [x] Import variables from local files - `n8n variables import --file <path>`
- [x] Set/update variable values - `n8n variables set <KEY> <VALUE>`
- [x] Delete variables - `n8n variables delete <ID>`

### Project Management

- [x] List projects - `n8n projects list`
- [x] Create new projects - `n8n projects create <NAME>`
- [ ] ~Transfer workflows between projects~

### Audit & Security

- [ ] Generate audit reports for workflows
- [x] Validate workflow files locally before upload - `n8n workflows validate`

### Configuration & Setup

- [x] Initialize local configuration - `n8n config init`
- [x] Set/update n8n instance URL - `n8n config set instance_url <url>`
- [x] Set/update API key - `n8n config set api_key <key>`
- [x] Configure default project/profile - `n8n config profiles use <name>`
- [ ] Enable verbose logging - use `--debug` flag (already supported)

## Technical Enhancements

- [x] Implement retry logic for API requests - exponential backoff via `RetryTransport`
- [x] Add support for multiple n8n instances (profiles) - `n8n config profiles add/list/use`, `--profile` flag
- [ ] Create workspace configuration for team collaboration
- [x] Add support for environment-specific variables
- [x] Add check-dirty feature to CI pipeline to detect uncommitted generated files
