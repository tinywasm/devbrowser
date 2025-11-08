package devbrowser

import (
	"fmt"
	"strings"
)

// ToolExecutor defines how a tool should be executed
type ToolExecutor func(args map[string]any, progress chan<- string)

// ToolMetadata provides MCP tool configuration metadata
type ToolMetadata struct {
	Name        string
	Description string
	Parameters  []ParameterMetadata
	Execute     ToolExecutor // Execution function
}

// ParameterMetadata describes a tool parameter
type ParameterMetadata struct {
	Name        string
	Description string
	Required    bool
	Type        string
	EnumValues  []string
	Default     any
}

// GetMCPToolsMetadata returns metadata for all DevBrowser MCP tools
func (b *DevBrowser) GetMCPToolsMetadata() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_open",
			Description: "Open Chrome development browser pointing to the local Go server to test the full-stack app (Go backend + WASM frontend).",
			Parameters:  []ParameterMetadata{},
			Execute: func(args map[string]any, progress chan<- string) {
				if b.isOpen {
					progress <- "Browser is already open"
					return
				}

				b.OpenBrowser()
				progress <- "Browser opened successfully"
			},
		},
		{
			Name:        "browser_close",
			Description: "Close Chrome development browser and cleanup resources when done testing or to restart fresh.",
			Parameters:  []ParameterMetadata{},
			Execute: func(args map[string]any, progress chan<- string) {
				if !b.isOpen {
					progress <- "Browser is already closed"
					return
				}

				if err := b.CloseBrowser(); err != nil {
					progress <- fmt.Sprintf("Failed to close browser: %v", err)
					return
				}
				progress <- "Browser closed successfully"
			},
		},
		{
			Name:        "browser_reload",
			Description: "Reload browser page to see latest WASM/asset changes without full browser restart (faster iteration during development).",
			Parameters:  []ParameterMetadata{},
			Execute: func(args map[string]any, progress chan<- string) {
				if !b.isOpen {
					progress <- "Browser is not open. Please open it first with browser_open"
					return
				}

				if err := b.Reload(); err != nil {
					progress <- fmt.Sprintf("Failed to reload browser: %v", err)
					return
				}
				progress <- "Browser reloaded successfully"
			},
		},
		{
			Name:        "browser_get_console",
			Description: "Get browser JavaScript console logs to debug WASM runtime errors, console.log outputs, or frontend issues. Filter by level (all/error/warning/log).",
			Parameters: []ParameterMetadata{
				{
					Name:        "level",
					Description: "Log level filter",
					Required:    false,
					Type:        "string",
					EnumValues:  []string{"all", "error", "warning", "log"},
					Default:     "all",
				},
				{
					Name:        "lines",
					Description: "Number of recent entries to retrieve",
					Required:    false,
					Type:        "number",
					Default:     50,
				},
			},
			Execute: func(args map[string]any, progress chan<- string) {
				if !b.isOpen {
					progress <- "Browser is not open. Please open it first with browser_open"
					return
				}

				logs, err := b.GetConsoleLogs()
				if err != nil {
					progress <- fmt.Sprintf("Failed to get console logs: %v", err)
					return
				}

				// b.logger("DEBUG:", logs)

				if len(logs) == 0 {
					progress <- "No console logs available"
					return
				}

				// Extract level filter
				level := "all"
				if levelValue, ok := args["level"]; ok {
					if levelStr, ok := levelValue.(string); ok {
						level = levelStr
					}
				}

				// Extract lines limit
				maxLines := 50
				if linesValue, ok := args["lines"]; ok {
					if linesFloat, ok := linesValue.(float64); ok {
						maxLines = int(linesFloat)
					}
				}

				// Filter logs by level
				var filteredLogs []string
				for _, log := range logs {
					if level == "all" {
						filteredLogs = append(filteredLogs, log)
					} else {
						upperLevel := strings.ToUpper(level)
						if strings.Contains(log, "["+upperLevel+"]") {
							filteredLogs = append(filteredLogs, log)
						}
					}
				}

				// Limit number of lines
				if len(filteredLogs) > maxLines {
					filteredLogs = filteredLogs[len(filteredLogs)-maxLines:]
				}

				if len(filteredLogs) == 0 {
					progress <- "No console logs"
					return
				}

				// Send logs directly without prefix (clean output like browser console)
				for _, log := range filteredLogs {
					progress <- log
				}
			},
		},
	}
}
