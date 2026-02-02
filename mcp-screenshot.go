package devbrowser

import (
	"fmt"

	"github.com/chromedp/chromedp"
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
				if !b.isOpen {
					b.Logger("Browser is not open. Please open it first with browser_open")
					return
				}

				fullpage := false
				if fp, ok := args["fullpage"].(bool); ok {
					fullpage = fp
				}

				var buf []byte
				var err error

				if fullpage {
					err = chromedp.Run(b.ctx,
						chromedp.FullScreenshot(&buf, 90),
					)
				} else {
					err = chromedp.Run(b.ctx,
						chromedp.CaptureScreenshot(&buf),
					)
				}

				if err != nil {
					b.Logger(fmt.Sprintf("Failed to capture screenshot: %v", err))
					return
				}
				if len(buf) == 0 {
					b.Logger("Screenshot capture returned empty buffer")
					return
				}

				// Write PNG image to clipboard
				clipboard.Write(clipboard.FmtImage, buf)
				b.Logger("Screenshot copied to clipboard")

				// Capture comprehensive page context for AI understanding (no OCR needed)
				var pageTitle, pageURL, htmlStructure string
				var windowWidth, windowHeight int

				err = chromedp.Run(b.ctx,
					chromedp.Title(&pageTitle),
					chromedp.Location(&pageURL),
					chromedp.Evaluate(`window.innerWidth`, &windowWidth),
					chromedp.Evaluate(`window.innerHeight`, &windowHeight),
					// Extract HTML structure with visible text and computed styles
					chromedp.Evaluate(getStructureJS, &htmlStructure),
				)

				// Build visual context report (what AI "sees" without image bytes)
				contextReport := fmt.Sprintf(
					"Screenshot captured (%d KB)\n"+
						"URL: %s | Title: %s | Viewport: %dx%d\n"+
						"\n"+
						"%s",
					len(buf)/1024,
					pageURL,
					pageTitle,
					windowWidth, windowHeight,
					htmlStructure,
				)

				if err != nil {
					// Fallback if context extraction fails
					contextReport = fmt.Sprintf("Screenshot captured (%d KB)", len(buf)/1024)
				}

				// Send both text context and binary data
				b.Logger(contextReport, mcpserve.BinaryData{MimeType: "image/png", Data: buf})
			},
		},
	}
}
