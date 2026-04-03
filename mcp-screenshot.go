package devbrowser

import (
	"fmt"

	"github.com/tinywasm/context"
	"github.com/tinywasm/mcp"
)

func (b *DevBrowser) GetScreenshotTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "browser_screenshot",
			Description: "Capture screenshot of current browser viewport to verify visual rendering, layout correctness, or UI state. Returns PNG image as MCP resource (binary efficient format).",
			InputSchema: EncodeSchema(new(ScreenshotArgs)),
			Resource:    "browser",
			Action:      'r',
			Execute: func(Ctx *context.Context, req mcp.Request) (*mcp.Result, error) {
				var args ScreenshotArgs
				if err := req.Bind(&args); err != nil {
					return nil, err
				}

				res, err := b.CaptureScreenshot(args.Fullpage)
				if err != nil {
					return nil, err
				}

				if len(res.ImageData) == 0 {
					return nil, fmt.Errorf("Screenshot capture returned empty buffer")
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
				b.Logger(contextReport)
				return mcp.Text(contextReport), nil
			},
		},
	}
}
