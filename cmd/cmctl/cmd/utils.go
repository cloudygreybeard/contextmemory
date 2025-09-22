package cmd

import (
	"fmt"
	"strings"
	"time"
)

// formatLabels formats labels for detailed display
func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return "<none>"
	}

	var pairs []string
	for k, v := range labels {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ",")
}

// formatLabelsCompact formats labels for compact table display
func formatLabelsCompact(labels map[string]string) string {
	if len(labels) == 0 {
		return "<none>"
	}

	// Show only first 2 labels for table display
	var pairs []string
	count := 0
	for k, v := range labels {
		if count >= 2 {
			pairs = append(pairs, "...")
			break
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
		count++
	}
	return strings.Join(pairs, ",")
}

// formatAge formats time duration in kubectl-style (e.g., "2d", "3h", "45m")
func formatAge(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	if duration < 30*24*time.Hour {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
	return fmt.Sprintf("%dw", int(duration.Hours()/(24*7)))
}

// truncateString truncates a string to maxLen with appropriate padding
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// parseLabels parses a comma-separated label selector string into a map
// Format: "key1=value1,key2=value2" -> map[string]string{"key1": "value1", "key2": "value2"}
func parseLabels(labelSelector string) map[string]string {
	labelMap := make(map[string]string)
	if labelSelector == "" {
		return labelMap
	}

	pairs := strings.Split(labelSelector, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" && value != "" {
				labelMap[key] = value
			}
		}
	}
	return labelMap
}
