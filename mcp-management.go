package devbrowser

import "fmt"

func (b *DevBrowser) getManagementTools() []ToolMetadata {
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
	}
}
