package n8n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MaxLimit is the maximum number of workflows that can be fetched per request (per n8n docs)
const MaxLimit = 250

// Client is a simple client for interacting with n8n API
type Client struct {
	baseURL  string
	apiToken string
	client   *http.Client
	logger   *zap.SugaredLogger
}

// NewClient creates a new n8n client
func NewClient(baseURL, apiToken string) *Client {
	var logger *zap.SugaredLogger

	if os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "true" {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		zapLogger, err := cfg.Build()
		if err != nil {
			zapLogger, _ = zap.NewProduction()
		}

		logger = zapLogger.Sugar().Named("n8n-api")
	} else {
		zapLogger, _ := zap.NewProduction()
		logger = zapLogger.Sugar().Named("n8n-api")
	}

	return &Client{
		baseURL:  baseURL + "/api/v1",
		apiToken: apiToken,
		client: &http.Client{
			Transport: &RetryTransport{
				Base:   http.DefaultTransport,
				Config: DefaultRetryConfig,
				Logger: logger,
			},
		},
		logger: logger,
	}
}

// NewClientWithConfig creates a new n8n client with a custom retry configuration
func NewClientWithConfig(baseURL, apiToken string, retryConfig RetryConfig) *Client {
	var logger *zap.SugaredLogger

	if os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "true" {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		zapLogger, err := cfg.Build()
		if err != nil {
			zapLogger, _ = zap.NewProduction()
		}

		logger = zapLogger.Sugar().Named("n8n-api")
	} else {
		zapLogger, _ := zap.NewProduction()
		logger = zapLogger.Sugar().Named("n8n-api")
	}

	return &Client{
		baseURL:  baseURL + "/api/v1",
		apiToken: apiToken,
		client: &http.Client{
			Transport: &RetryTransport{
				Base:   http.DefaultTransport,
				Config: retryConfig,
				Logger: logger,
			},
		},
		logger: logger,
	}
}

// logDebug logs a debug message
func (c *Client) logDebug(format string, args ...interface{}) {
	c.logger.Debugf(format, args...)
}

// WorkflowFilters holds optional filters for GetWorkflows
type WorkflowFilters struct {
	Active    *bool
	Tags      string
	Name      string
	ProjectID string
}

// GetWorkflows fetches workflows from the n8n API
// If limit is nil, uses the API's default (100)
// If limit is provided, returns up to that many workflows (max MaxLimit)
// If filters is provided, applies the given filters as query parameters
func (c *Client) GetWorkflows(limit *int, filters *WorkflowFilters) (*WorkflowList, error) {
	apiURL := fmt.Sprintf("%s/workflows", c.baseURL)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if limit != nil {
		requestLimit := min(*limit, MaxLimit)
		q.Add("limit", strconv.Itoa(requestLimit))
	}
	if filters != nil {
		if filters.Active != nil {
			if *filters.Active {
				q.Add("active", "true")
			} else {
				q.Add("active", "false")
			}
		}
		if filters.Tags != "" {
			q.Add("tags", filters.Tags)
		}
		if filters.Name != "" {
			q.Add("name", filters.Name)
		}
		if filters.ProjectID != "" {
			q.Add("projectId", filters.ProjectID)
		}
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result WorkflowList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ActivateWorkflow activates a workflow by ID
func (c *Client) ActivateWorkflow(id string) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s/activate", c.baseURL, id)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result Workflow
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeactivateWorkflow deactivates a workflow by ID
func (c *Client) DeactivateWorkflow(id string) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s/deactivate", c.baseURL, id)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result Workflow
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateWorkflow creates a new workflow
func (c *Client) CreateWorkflow(workflow *Workflow) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows", c.baseURL)

	workflowCopy := *workflow
	workflowCopy.Id = nil
	workflowCopy.Active = nil
	workflowCopy.CreatedAt = nil
	workflowCopy.UpdatedAt = nil
	workflowCopy.Tags = nil

	body, err := json.Marshal(workflowCopy)
	if err != nil {
		return nil, fmt.Errorf("error marshaling workflow: %w", err)
	}

	c.logDebug("CREATE WORKFLOW REQUEST: %s", string(body))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		c.logDebug("CREATE WORKFLOW FORMATTED JSON:\n%s", prettyJSON.String())
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	c.logDebug("CREATE/UPDATE WORKFLOW RESPONSE (Status: %d): %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, respBody)
	}

	var w Workflow
	if err := json.NewDecoder(bytes.NewBuffer(respBody)).Decode(&w); err != nil {
		return nil, err
	}

	return &w, nil
}

// UpdateWorkflow updates an existing workflow by its ID
func (c *Client) UpdateWorkflow(id string, workflow *Workflow) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.baseURL, id)

	workflowCopy := *workflow
	workflowCopy.Id = nil
	workflowCopy.Active = nil
	workflowCopy.CreatedAt = nil
	workflowCopy.UpdatedAt = nil
	workflowCopy.Tags = nil

	body, err := json.Marshal(workflowCopy)
	if err != nil {
		return nil, fmt.Errorf("error marshaling workflow: %w", err)
	}

	c.logDebug("UPDATE WORKFLOW REQUEST (ID: %s): %s", id, string(body))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		c.logDebug("UPDATE WORKFLOW FORMATTED JSON (ID: %s):\n%s", id, prettyJSON.String())
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var w Workflow
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, err
	}

	return &w, nil
}

// GetWorkflow fetches a single workflow by its ID
func (c *Client) GetWorkflow(id string) (*Workflow, error) {
	url := fmt.Sprintf("%s/workflows/%s", c.baseURL, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var workflow Workflow
	if err := json.NewDecoder(resp.Body).Decode(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// DeleteWorkflow deletes a workflow by ID
func (c *Client) DeleteWorkflow(id string) error {
	url := fmt.Sprintf("%s/workflows/%s", c.baseURL, id)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}

// GetExecutions fetches workflow executions from the n8n API
// workflowID is optional - if provided, only executions for that workflow will be returned
// includeData is optional - if provided as true, execution data will be included in the response
// status is optional - if provided, only executions with that status will be returned (error, success, waiting)
// limit is optional - if provided, limits the number of executions returned
// cursor is optional - if provided, retrieves the next page of results
func (c *Client) GetExecutions(workflowID string, includeData bool, status string, limit int, cursor string) (*ExecutionList, error) {
	baseURL := fmt.Sprintf("%s/executions", c.baseURL)

	params := url.Values{}
	if workflowID != "" {
		params.Add("workflowId", workflowID)
	}
	if includeData {
		params.Add("includeData", "true")
	}
	if status != "" {
		params.Add("status", status)
	}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}
	if cursor != "" {
		params.Add("cursor", cursor)
	}

	requestURL := baseURL
	if len(params) > 0 {
		requestURL = fmt.Sprintf("%s?%s", baseURL, params.Encode())
	}

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var flexibleResult ExecutionListWithFlexibleIDs
	if err := json.NewDecoder(resp.Body).Decode(&flexibleResult); err != nil {
		return nil, fmt.Errorf("failed to decode execution list: %v", err)
	}

	result := flexibleResult.ToExecutionList()
	return result, nil
}

// GetExecutionById fetches a specific execution by its ID
// includeData is optional - if provided as true, execution data will be included in the response
func (c *Client) GetExecutionById(executionID string, includeData bool) (*Execution, error) {
	baseURL := fmt.Sprintf("%s/executions/%s", c.baseURL, executionID)

	params := url.Values{}
	if includeData {
		params.Add("includeData", "true")
	}

	requestURL := baseURL
	if len(params) > 0 {
		requestURL = fmt.Sprintf("%s?%s", baseURL, params.Encode())
	}

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var flexibleResult ExecutionWithFlexibleIDs
	if err := json.NewDecoder(resp.Body).Decode(&flexibleResult); err != nil {
		return nil, fmt.Errorf("failed to decode execution: %v", err)
	}

	result := toExecution(flexibleResult)
	return &result, nil
}

// GetWorkflowTags fetches the tags of a workflow by its ID
func (c *Client) GetWorkflowTags(id string) (WorkflowTags, error) {
	url := fmt.Sprintf("%s/workflows/%s/tags", c.baseURL, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var tags WorkflowTags
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// UpdateWorkflowTags updates the tags of a workflow by its ID
func (c *Client) UpdateWorkflowTags(id string, tagIds TagIds) (WorkflowTags, error) {
	url := fmt.Sprintf("%s/workflows/%s/tags", c.baseURL, id)

	jsonBody, err := json.Marshal(tagIds)
	if err != nil {
		return nil, err
	}

	c.logDebug("UPDATE WORKFLOW TAGS REQUEST (ID: %s): %s", id, string(jsonBody))

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, jsonBody, "", "  "); err == nil {
		c.logDebug("UPDATE WORKFLOW TAGS FORMATTED JSON (ID: %s):\n%s", id, prettyJSON.String())
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var tags WorkflowTags
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// CreateTag creates a new tag in n8n
func (c *Client) CreateTag(tagName string) (*Tag, error) {
	url := fmt.Sprintf("%s/tags", c.baseURL)

	tagRequest := map[string]string{"name": tagName}
	jsonBody, err := json.Marshal(tagRequest)
	if err != nil {
		return nil, err
	}

	c.logDebug("CREATE TAG REQUEST: %s", string(jsonBody))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var tag Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

// GetTags fetches all tags from n8n
func (c *Client) GetTags() (*TagList, error) {
	url := fmt.Sprintf("%s/tags", c.baseURL)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result TagList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetVariables fetches environment variables from the n8n API
func (c *Client) GetVariables(limit *int, cursor string) (*VariableList, error) {
	apiURL := fmt.Sprintf("%s/variables", c.baseURL)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if limit != nil {
		requestLimit := min(*limit, MaxLimit)
		q.Add("limit", strconv.Itoa(requestLimit))
	}
	if cursor != "" {
		q.Add("cursor", cursor)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result VariableList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateVariable creates a new environment variable
func (c *Client) CreateVariable(variable *Variable) error {
	apiURL := fmt.Sprintf("%s/variables", c.baseURL)

	body, err := json.Marshal(variable)
	if err != nil {
		return fmt.Errorf("error marshaling variable: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}

// UpdateVariable updates an existing environment variable by its ID
func (c *Client) UpdateVariable(id string, variable *Variable) error {
	apiURL := fmt.Sprintf("%s/variables/%s", c.baseURL, id)

	body, err := json.Marshal(variable)
	if err != nil {
		return fmt.Errorf("error marshaling variable: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}

// DeleteVariable deletes an environment variable by its ID
func (c *Client) DeleteVariable(id string) error {
	apiURL := fmt.Sprintf("%s/variables/%s", c.baseURL, id)

	req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}

// GetProjects fetches projects from the n8n API
func (c *Client) GetProjects(limit *int, cursor string) (*ProjectList, error) {
	apiURL := fmt.Sprintf("%s/projects", c.baseURL)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if limit != nil {
		requestLimit := min(*limit, MaxLimit)
		q.Add("limit", strconv.Itoa(requestLimit))
	}
	if cursor != "" {
		q.Add("cursor", cursor)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result ProjectList
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(name string) error {
	apiURL := fmt.Sprintf("%s/projects", c.baseURL)

	project := Project{Name: name}
	body, err := json.Marshal(project)
	if err != nil {
		return fmt.Errorf("error marshaling project: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}

// GetCredentialSchema fetches the schema for a credential type
func (c *Client) GetCredentialSchema(credentialTypeName string) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/credentials/schema/%s", c.baseURL, credentialTypeName)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateCredential creates a new credential
func (c *Client) CreateCredential(credential *Credential) (*CreateCredentialResponse, error) {
	apiURL := fmt.Sprintf("%s/credentials", c.baseURL)

	body, err := json.Marshal(credential)
	if err != nil {
		return nil, fmt.Errorf("error marshaling credential: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result CreateCredentialResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteCredential deletes a credential by its ID
func (c *Client) DeleteCredential(id string) error {
	apiURL := fmt.Sprintf("%s/credentials/%s", c.baseURL, id)

	req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}

// RetryExecution retries a failed execution
func (c *Client) RetryExecution(executionID string, loadWorkflow bool) (*Execution, error) {
	apiURL := fmt.Sprintf("%s/executions/%s/retry", c.baseURL, executionID)

	reqBody := map[string]bool{"loadWorkflow": loadWorkflow}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var flexibleResult ExecutionWithFlexibleIDs
	if err := json.NewDecoder(resp.Body).Decode(&flexibleResult); err != nil {
		return nil, fmt.Errorf("failed to decode execution: %v", err)
	}

	result := toExecution(flexibleResult)
	return &result, nil
}

// GenerateAudit generates a security audit report for the n8n instance
func (c *Client) GenerateAudit(options *PostAuditJSONBody) (*Audit, error) {
	apiURL := fmt.Sprintf("%s/audit", c.baseURL)

	var body []byte
	var err error
	if options != nil {
		body, err = json.Marshal(options)
		if err != nil {
			return nil, fmt.Errorf("error marshaling audit options: %w", err)
		}
	} else {
		body = []byte("{}")
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	var result Audit
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteExecution deletes an execution by its ID
func (c *Client) DeleteExecution(executionID string) error {
	apiURL := fmt.Sprintf("%s/executions/%s", c.baseURL, executionID)

	req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-N8N-API-KEY", c.apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Warnf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned error %d: %s", resp.StatusCode, body)
	}

	return nil
}
