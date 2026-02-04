package devbrowser

import (
	"fmt"

	"github.com/tinywasm/mcpserve"
	"golang.design/x/clipboard"
)

func (b *DevBrowser) getScreenshotTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_screenshot",
			Description: "Capture screenshot of current browser viewport to verify visual rendering, layout correctness, or UI state. Returns PNG image as MCP resource (binary efficient format).",
			Parameters: []ParameterMetadata{
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
				clipboard.Write(clipboard.FmtImage, res.ImageData)
				b.Logger("Screenshot copied to clipboard")

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
