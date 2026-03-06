package devbrowser

import (
	"fmt"

	"github.com/tinywasm/mcp"
)

func (b *DevBrowser) getNavigationTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "browser_navigate",
			Description: "Navigate the browser to a specific URL",
			Parameters: []mcp.Parameter{
				{
					Name:        "url",
					Description: "Complete URL (including http:// or https://)",
					Required:    true,
					Type:        "string",
				},
			},
			Execute: func(args map[string]any) {
				if !b.isOpen {
					b.Logger("Browser is not open. Please open it first with browser_open")
					return
				}

				url, ok := args["url"].(string)
				if !ok || url == "" {
					b.Logger("URL parameter is required")
					return
				}

				err := b.navigateToURL(url)
				if err != nil {
					b.Logger(fmt.Sprintf("Error navigating to %s: %v", url, err))
					return
				}

				b.Logger(fmt.Sprintf("Navigated to %s", url))
			},
		},
	}
}
