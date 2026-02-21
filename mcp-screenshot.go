package devbrowser

import (
	"fmt"

	"github.com/tinywasm/mcpserve"
)

func (b *DevBrowser) getScreenshotTools() []mcpserve.ToolMetadata {
	return []mcpserve.ToolMetadata{
		{
			Name:        "browser_screenshot",
			Description: "Capture screenshot of current browser viewport to verify visual rendering, layout correctness, or UI state. Returns PNG image as MCP resource (binary efficient format).",
			Parameters: []mcpserve.ParameterMetadata{
				{
					Name:        "fullpage",
					Description: "Capture full page height instead of viewport only",
					Required:    false,
					Type:        "boolean",
					Default:     false,
				},
			},
			Execute: func(args map[string]any) {
				fullpage := false
				if fp, ok := args["fullpage"].(bool); ok {
					fullpage = fp
				}

				res, err := b.CaptureScreenshot(fullpage)
				if err != nil {
					b.Logger(err.Error())
					return
				}

				if len(res.ImageData) == 0 {
					b.Logger("Screenshot capture returned empty buffer")
					return
				}

				// Write PNG image to clipboard
				if err := writeToClipboard(res.ImageData); err == nil {
					b.Logger("Screenshot copied to clipboard")
				} else {
					b.Logger(fmt.Sprintf("Failed to copy screenshot to clipboard: %v", err))
				}

				// Build visual context report (what AI "sees" without image bytes)
				contextReport := fmt.Sprintf(
					"Screenshot captured (%d KB)\n"+
						"URL: %s | Title: %s | Viewport: %dx%d\n"+
						"\n"+
						"%s",
					len(res.ImageData)/1024,
					res.PageURL,
					res.PageTitle,
					res.Width, res.Height,
					res.HTMLStructure,
				)

				// Send both text context and binary data
				b.Logger(contextReport, mcpserve.BinaryData{MimeType: "image/png", Data: res.ImageData})
			},
		},
	}
}
