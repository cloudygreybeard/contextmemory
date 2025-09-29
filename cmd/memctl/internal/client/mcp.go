package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *MCPError) Error() string {
	return fmt.Sprintf("MCP Error %d: %s", e.Code, e.Message)
}

// MCPClient handles communication with the MCP server
type MCPClient struct {
	serverPath string
	verbosity  int
	requestID  int
	mutex      sync.Mutex
}

// NewMCPClient creates a new MCP client
func NewMCPClient(serverPath string, verbosity int) *MCPClient {
	return &MCPClient{
		serverPath: serverPath,
		verbosity:  verbosity,
		requestID:  0,
	}
}

// FindMCPServer attempts to locate the MCP server binary
func FindMCPServer(providedPath string) (string, error) {
	if providedPath != "" {
		if _, err := os.Stat(providedPath); err == nil {
			return providedPath, nil
		}
		return "", fmt.Errorf("MCP server not found at specified path: %s", providedPath)
	}

	// Search in PATH
	if path, err := exec.LookPath("mcp-server"); err == nil {
		return path, nil
	}

	// Search relative to current directory (for development)
	cwd, err := os.Getwd()
	if err == nil {
		candidates := []string{
			filepath.Join(cwd, "build/mcp-server"),
			filepath.Join(cwd, "../build/mcp-server"),
			filepath.Join(cwd, "../../build/mcp-server"),
			filepath.Join(cwd, "mcp-server"),
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("MCP server not found. Please specify path with --mcp-server flag or ensure mcp-server is in PATH")
}

// Call makes a JSON-RPC call to the MCP server
func (c *MCPClient) Call(method string, params interface{}) (*MCPResponse, error) {
	c.mutex.Lock()
	c.requestID++
	reqID := c.requestID
	c.mutex.Unlock()

	// Create request
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params:  params,
	}

	// Marshal request
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.verbosity >= 2 {
		fmt.Fprintf(os.Stderr, "MCP Request: %s\n", string(reqData))
	}

	// Start MCP server process
	cmd := exec.Command(c.serverPath)
	cmd.Stdin = bytes.NewReader(reqData)
	cmd.Stderr = os.Stderr

	// Get response
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to communicate with MCP server: %w", err)
	}

	if c.verbosity >= 2 {
		fmt.Fprintf(os.Stderr, "MCP Response: %s\n", string(output))
	}

	// Parse response
	var resp MCPResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse MCP response: %w", err)
	}

	// Check for errors
	if resp.Error != nil {
		return &resp, resp.Error
	}

	return &resp, nil
}

// Memory-specific convenience methods

// ListMemories calls memories/v1/list
func (c *MCPClient) ListMemories(namespace string, labelSelector map[string]string, includeContent bool, limit int) (*MCPResponse, error) {
	params := map[string]interface{}{
		"namespace":      namespace,
		"includeContent": includeContent,
	}

	if labelSelector != nil && len(labelSelector) > 0 {
		params["labelSelector"] = labelSelector
	}

	if limit > 0 {
		params["limit"] = limit
	}

	return c.Call("memories/v1/list", params)
}

// GetMemory calls memories/v1/get
func (c *MCPClient) GetMemory(namespace, name string) (*MCPResponse, error) {
	params := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	}

	return c.Call("memories/v1/get", params)
}

// CreateMemory calls memories/v1/create
func (c *MCPClient) CreateMemory(namespace, name, content string, labels map[string]string) (*MCPResponse, error) {
	memory := map[string]interface{}{
		"name":    name,
		"content": content,
	}

	if labels != nil && len(labels) > 0 {
		memory["labels"] = labels
	}

	params := map[string]interface{}{
		"namespace": namespace,
		"memory":    memory,
	}

	return c.Call("memories/v1/create", params)
}

// UpdateMemory calls memories/v1/update
func (c *MCPClient) UpdateMemory(namespace, name, newName, content string, labels map[string]string) (*MCPResponse, error) {
	memory := map[string]interface{}{}

	if newName != "" {
		memory["name"] = newName
	}
	if content != "" {
		memory["content"] = content
	}
	if labels != nil {
		memory["labels"] = labels
	}

	params := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"memory":    memory,
	}

	return c.Call("memories/v1/update", params)
}

// PatchMemory calls memories/v1/patch
func (c *MCPClient) PatchMemory(namespace, name string, patch map[string]interface{}) (*MCPResponse, error) {
	params := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
		"patch":     patch,
	}

	return c.Call("memories/v1/patch", params)
}

// DeleteMemory calls memories/v1/delete
func (c *MCPClient) DeleteMemory(namespace, name string) (*MCPResponse, error) {
	params := map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	}

	return c.Call("memories/v1/delete", params)
}

// SearchMemories calls memories/v1/search
func (c *MCPClient) SearchMemories(query, namespace string, labelSelector map[string]string, includeContent bool, limit int) (*MCPResponse, error) {
	params := map[string]interface{}{
		"namespace":      namespace,
		"includeContent": includeContent,
	}

	if query != "" {
		params["query"] = query
	}

	if labelSelector != nil && len(labelSelector) > 0 {
		params["labelSelector"] = labelSelector
	}

	if limit > 0 {
		params["limit"] = limit
	}

	return c.Call("memories/v1/search", params)
}

// Utility functions

// ParseLabels parses a label string in format "key1=value1,key2=value2"
func ParseLabels(labelsStr string) map[string]string {
	if labelsStr == "" {
		return nil
	}

	labels := make(map[string]string)
	pairs := strings.Split(labelsStr, ",")

	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			labels[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return labels
}

// ReadStdin reads content from stdin
func ReadStdin() (string, error) {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice == 0 {
		// Data is being piped
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(data), nil
	}
	return "", nil
}
