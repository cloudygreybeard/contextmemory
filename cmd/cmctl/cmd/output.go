package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/cloudygreybeard/contextmemory/cmd/cmctl/internal/storage"
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

// FormatMemoryList formats a list of memories according to output options
func FormatMemoryList(memories []storage.Memory, opts OutputOptions, showID bool) (string, error) {
	switch opts.Format {
	case OutputFormatTable:
		return formatMemoryTable(memories, showID), nil
	case OutputFormatJSON, OutputFormatYAML, OutputFormatJSONPath, OutputFormatGoTemplate:
		// Create a wrapper structure for consistent API output
		output := struct {
			APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
			Kind       string           `json:"kind" yaml:"kind"`
			Items      []storage.Memory `json:"items" yaml:"items"`
		}{
			APIVersion: "contextmemory.io/v1",
			Kind:       "MemoryList",
			Items:      memories,
		}
		return FormatOutput(output, opts)
	default:
		return "", fmt.Errorf("unsupported output format: %s", opts.Format)
	}
}

// FormatSingleMemory formats a single memory according to output options
func FormatSingleMemory(memory *storage.Memory, opts OutputOptions) (string, error) {
	switch opts.Format {
	case OutputFormatTable:
		return formatSingleMemoryTable(memory), nil
	case OutputFormatJSON, OutputFormatYAML, OutputFormatJSONPath, OutputFormatGoTemplate:
		// Create a wrapper structure for consistent API output
		output := struct {
			APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
			Kind       string         `json:"kind" yaml:"kind"`
			Metadata   map[string]any `json:"metadata" yaml:"metadata"`
			Spec       storage.Memory `json:"spec" yaml:"spec"`
		}{
			APIVersion: "contextmemory.io/v1",
			Kind:       "Memory",
			Metadata: map[string]any{
				"id":   memory.ID,
				"name": memory.Name,
			},
			Spec: *memory,
		}
		return FormatOutput(output, opts)
	default:
		return "", fmt.Errorf("unsupported output format: %s", opts.Format)
	}
}

// formatMemoryTable formats memories as a table (existing logic)
func formatMemoryTable(memories []storage.Memory, showID bool) string {
	if len(memories) == 0 {
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
	for _, memory := range memories {
		labels := formatLabelsCompact(memory.Labels)
		age := formatAge(memory.UpdatedAt)

		if showID {
			result.WriteString(fmt.Sprintf("%-24s %-32s %-26s %-20s\n",
				truncateString(memory.ID, 22),
				truncateString(memory.Name, 30),
				truncateString(labels, 24),
				age))
		} else {
			result.WriteString(fmt.Sprintf("%-40s %-30s %-20s\n",
				truncateString(memory.Name, 38),
				truncateString(labels, 28),
				age))
		}
	}

	return result.String()
}

// formatSingleMemoryTable formats a single memory as table
func formatSingleMemoryTable(memory *storage.Memory) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Name:\t%s\n", memory.Name))
	result.WriteString(fmt.Sprintf("ID:\t%s\n", memory.ID))
	result.WriteString(fmt.Sprintf("Created:\t%s\n", memory.CreatedAt.Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("Updated:\t%s\n", memory.UpdatedAt.Format("2006-01-02 15:04:05")))

	if len(memory.Labels) > 0 {
		result.WriteString("Labels:\t")
		labels := make([]string, 0, len(memory.Labels))
		for key, value := range memory.Labels {
			labels = append(labels, fmt.Sprintf("%s=%s", key, value))
		}
		result.WriteString(strings.Join(labels, ","))
		result.WriteString("\n")
	} else {
		result.WriteString("Labels:\tnone\n")
	}

	result.WriteString("\nContent:\n")
	result.WriteString(memory.Content)
	result.WriteString("\n")

	return result.String()
}
