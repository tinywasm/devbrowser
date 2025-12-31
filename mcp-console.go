package devbrowser

import (
	"fmt"
	"strings"
)

func (b *DevBrowser) getConsoleTools() []ToolMetadata {
	return []ToolMetadata{
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
			Execute: func(args map[string]any) {
				if !b.isOpen {
					b.Logger("Browser is not open. Please open it first with browser_open")
					return
				}

				logs, err := b.GetConsoleLogs()
				if err != nil {
					b.Logger(fmt.Sprintf("Failed to get console logs: %v", err))
					return
				}

				if len(logs) == 0 {
					b.Logger("No console logs available")
					return
				}

				level := "all"
				if levelValue, ok := args["level"]; ok {
					if levelStr, ok := levelValue.(string); ok {
						level = levelStr
					}
				}

				maxLines := 50
				if linesValue, ok := args["lines"]; ok {
					if linesFloat, ok := linesValue.(float64); ok {
						maxLines = int(linesFloat)
					}
				}

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

				if len(filteredLogs) > maxLines {
					filteredLogs = filteredLogs[len(filteredLogs)-maxLines:]
				}

				if len(filteredLogs) == 0 {
					b.Logger("No console logs")
					return
				}

				for _, log := range filteredLogs {
					b.Logger(log)
				}
			},
		},
	}
}
