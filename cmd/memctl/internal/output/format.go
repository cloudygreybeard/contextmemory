package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/jsonpath"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	OutputFormatTable      OutputFormat = "table"
	OutputFormatJSON       OutputFormat = "json"
	OutputFormatYAML       OutputFormat = "yaml"
	OutputFormatJSONPath   OutputFormat = "jsonpath"
	OutputFormatGoTemplate OutputFormat = "go-template"
)

// OutputOptions contains options for formatting output
type OutputOptions struct {
	Format   OutputFormat
	Template string // For jsonpath or go-template
}

// FormatOutput formats the given data according to the output options
func FormatOutput(data interface{}, opts OutputOptions) (string, error) {
	switch opts.Format {
	case OutputFormatJSON:
		return formatJSON(data)
	case OutputFormatYAML:
		return formatYAML(data)
	case OutputFormatJSONPath:
		return formatJSONPath(data, opts.Template)
	case OutputFormatGoTemplate:
		return formatGoTemplate(data, opts.Template)
	case OutputFormatTable:
		fallthrough
	default:
		return "", fmt.Errorf("table format should be handled by the calling function")
	}
}

// formatJSON formats data as JSON
func formatJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData), nil
}

// formatYAML formats data as YAML
func formatYAML(data interface{}) (string, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(yamlData), nil
}

// formatJSONPath formats data using JSONPath template
func formatJSONPath(data interface{}, template string) (string, error) {
	if template == "" {
		return "", fmt.Errorf("jsonpath template is required")
	}

	// Convert data to JSON first
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data for jsonpath: %w", err)
	}

	// Parse the JSONPath template
	jp := jsonpath.New("output")
	if err := jp.Parse(template); err != nil {
		return "", fmt.Errorf("failed to parse jsonpath template: %w", err)
	}

	// Apply the template to the data
	var obj interface{}
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return "", fmt.Errorf("failed to unmarshal data for jsonpath: %w", err)
	}

	var buf bytes.Buffer
	if err := jp.Execute(&buf, obj); err != nil {
		return "", fmt.Errorf("failed to execute jsonpath template: %w", err)
	}

	return buf.String(), nil
}

// formatGoTemplate formats data using Go template
func formatGoTemplate(data interface{}, templateStr string) (string, error) {
	if templateStr == "" {
		return "", fmt.Errorf("go template is required")
	}

	// Parse the template
	tmpl, err := template.New("output").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse go template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute go template: %w", err)
	}

	return buf.String(), nil
}

// ParseOutputFormat parses the output format string
func ParseOutputFormat(format string) (OutputOptions, error) {
	// Handle formats like "jsonpath=.items[*].metadata.name" or "go-template={{.name}}"
	if strings.Contains(format, "=") {
		parts := strings.SplitN(format, "=", 2)
		if len(parts) != 2 {
			return OutputOptions{}, fmt.Errorf("invalid output format: %s", format)
		}

		formatType := parts[0]
		template := parts[1]

		switch formatType {
		case "jsonpath":
			return OutputOptions{Format: OutputFormatJSONPath, Template: template}, nil
		case "go-template":
			return OutputOptions{Format: OutputFormatGoTemplate, Template: template}, nil
		default:
			return OutputOptions{}, fmt.Errorf("unknown output format: %s", formatType)
		}
	}

	// Handle simple formats
	switch format {
	case "json":
		return OutputOptions{Format: OutputFormatJSON}, nil
	case "yaml":
		return OutputOptions{Format: OutputFormatYAML}, nil
	case "table", "":
		return OutputOptions{Format: OutputFormatTable}, nil
	default:
		return OutputOptions{}, fmt.Errorf("unknown output format: %s", format)
	}
}

// Memory-specific formatting functions

// MemoryResource represents a Kubernetes-style Memory resource
type MemoryResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
}

// MemoryListResource represents a list of Memory resources
type MemoryListResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Items      []MemoryResource       `json:"items"`
}

// FormatMemoryList formats a MemoryList according to output options
func FormatMemoryList(data interface{}, opts OutputOptions, showID bool) (string, error) {
	switch opts.Format {
	case OutputFormatTable:
		return formatMemoryTable(data, showID), nil
	case OutputFormatJSON, OutputFormatYAML, OutputFormatJSONPath, OutputFormatGoTemplate:
		return FormatOutput(data, opts)
	default:
		return "", fmt.Errorf("unsupported output format: %s", opts.Format)
	}
}

// FormatSingleMemory formats a single Memory resource according to output options
func FormatSingleMemory(data interface{}, opts OutputOptions) (string, error) {
	switch opts.Format {
	case OutputFormatTable:
		return formatSingleMemoryTable(data), nil
	case OutputFormatJSON, OutputFormatYAML, OutputFormatJSONPath, OutputFormatGoTemplate:
		return FormatOutput(data, opts)
	default:
		return "", fmt.Errorf("unsupported output format: %s", opts.Format)
	}
}

// formatMemoryTable formats memories as a table
func formatMemoryTable(data interface{}, showID bool) string {
	// Parse the response to extract items
	var items []MemoryResource

	dataBytes, _ := json.Marshal(data)

	// Try to parse as a direct MCP JSON-RPC response first (most common case)
	var response map[string]interface{}
	if err := json.Unmarshal(dataBytes, &response); err == nil {
		if result, ok := response["result"].(map[string]interface{}); ok {
			if itemsInterface, ok := result["items"].([]interface{}); ok {
				for _, item := range itemsInterface {
					if itemBytes, err := json.Marshal(item); err == nil {
						var memory MemoryResource
						if err := json.Unmarshal(itemBytes, &memory); err == nil {
							items = append(items, memory)
						}
					}
				}
			}
		}
	}

	// If that didn't work, try parsing as MemoryListResource directly
	if len(items) == 0 {
		var listResp MemoryListResource
		if err := json.Unmarshal(dataBytes, &listResp); err == nil {
			items = listResp.Items
		}
	}

	if len(items) == 0 {
		return "No resources found."
	}

	var result strings.Builder

	// Print header with conditional ID column
	if showID {
		result.WriteString(fmt.Sprintf("%-24s %-32s %-26s %-20s\n", "ID", "NAME", "LABELS", "AGE"))
	} else {
		result.WriteString(fmt.Sprintf("%-40s %-30s %-20s\n", "NAME", "LABELS", "AGE"))
	}

	// Print memories with conditional ID column
	for _, memory := range items {
		name := getStringFromMap(memory.Metadata, "name", "unknown")
		uid := getStringFromMap(memory.Metadata, "uid", "unknown")
		labels := formatLabels(memory.Metadata["labels"])
		age := formatAge(memory.Metadata["creationTimestamp"])

		if showID {
			result.WriteString(fmt.Sprintf("%-24s %-32s %-26s %-20s\n",
				truncateString(uid, 22),
				truncateString(name, 30),
				truncateString(labels, 24),
				age))
		} else {
			result.WriteString(fmt.Sprintf("%-40s %-30s %-20s\n",
				truncateString(name, 38),
				truncateString(labels, 28),
				age))
		}
	}

	return result.String()
}

// formatSingleMemoryTable formats a single memory as table
func formatSingleMemoryTable(data interface{}) string {
	// Parse the response
	var memory MemoryResource
	dataBytes, _ := json.Marshal(data)

	// Try to parse as MCP response first
	var response map[string]interface{}
	if err := json.Unmarshal(dataBytes, &response); err == nil {
		if result, ok := response["result"].(map[string]interface{}); ok {
			if resultBytes, err := json.Marshal(result); err == nil {
				json.Unmarshal(resultBytes, &memory)
			}
		}
	} else {
		// Try to parse directly
		json.Unmarshal(dataBytes, &memory)
	}

	var result strings.Builder

	name := getStringFromMap(memory.Metadata, "name", "unknown")
	uid := getStringFromMap(memory.Metadata, "uid", "unknown")
	namespace := getStringFromMap(memory.Metadata, "namespace", "default")
	created := memory.Metadata["creationTimestamp"]
	labels := memory.Metadata["labels"]
	content := getStringFromMap(memory.Spec, "content", "")

	result.WriteString(fmt.Sprintf("Name:\t%s\n", name))
	result.WriteString(fmt.Sprintf("Namespace:\t%s\n", namespace))
	result.WriteString(fmt.Sprintf("ID:\t%s\n", uid))
	result.WriteString(fmt.Sprintf("Created:\t%s\n", formatTimestamp(created)))

	if labels != nil {
		result.WriteString("Labels:\t")
		result.WriteString(formatLabels(labels))
		result.WriteString("\n")
	} else {
		result.WriteString("Labels:\tnone\n")
	}

	result.WriteString("\nContent:\n")
	result.WriteString(content)
	result.WriteString("\n")

	return result.String()
}

// Utility functions

func getStringFromMap(m map[string]interface{}, key, defaultVal string) string {
	if m == nil {
		return defaultVal
	}
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

func formatLabels(labelsInterface interface{}) string {
	if labelsInterface == nil {
		return "none"
	}

	labels, ok := labelsInterface.(map[string]interface{})
	if !ok {
		return "none"
	}

	if len(labels) == 0 {
		return "none"
	}

	var parts []string
	for k, v := range labels {
		if strVal, ok := v.(string); ok {
			parts = append(parts, fmt.Sprintf("%s=%s", k, strVal))
		}
	}

	return strings.Join(parts, ",")
}

func formatAge(timestampInterface interface{}) string {
	if timestampInterface == nil {
		return "unknown"
	}

	timestampStr, ok := timestampInterface.(string)
	if !ok {
		return "unknown"
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return "unknown"
	}

	duration := time.Since(timestamp)
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

func formatTimestamp(timestampInterface interface{}) string {
	if timestampInterface == nil {
		return "unknown"
	}

	timestampStr, ok := timestampInterface.(string)
	if !ok {
		return "unknown"
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return timestampStr
	}

	return timestamp.Format("2006-01-02 15:04:05")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
