package devbrowser

import "github.com/tinywasm/mcpserve"

import (
	"fmt"
)

func (b *DevBrowser) getConsoleTools() []mcpserve.ToolMetadata {
	return []mcpserve.ToolMetadata{
		{
			Name:        "browser_get_console",
			Description: "Get browser JavaScript console logs to debug WASM runtime errors, console.log outputs, or frontend issues.",
			Parameters: []mcpserve.ParameterMetadata{
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

				maxLines := 50
				if linesValue, ok := args["lines"]; ok {
					if linesFloat, ok := linesValue.(float64); ok {
						maxLines = int(linesFloat)
					}
				}

				if len(logs) > maxLines {
					logs = logs[len(logs)-maxLines:]
				}

				for _, log := range logs {
					b.Logger(log)
				}
			},
		},
	}
}
