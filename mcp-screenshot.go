package devbrowser

import (
	"encoding/base64"
	"fmt"
	"github.com/chromedp/chromedp"
)

func (b *DevBrowser) getScreenshotTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_screenshot",
			Description: "Capture screenshot of current browser viewport to verify visual rendering, layout correctness, or UI state. Returns base64-encoded PNG image.",
			Parameters: []ParameterMetadata{
				{
					Name:        "fullpage",
					Description: "Capture full page height instead of viewport only",
					Required:    false,
					Type:        "boolean",
					Default:     false,
				},
			},
			Execute: func(args map[string]any, progress chan<- string) {
				if !b.isOpen {
					progress <- "Browser is not open. Please open it first with browser_open"
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
					progress <- fmt.Sprintf("Failed to capture screenshot: %v", err)
					return
				}
				if len(buf) == 0 {
					progress <- "Screenshot capture returned empty buffer"
					return
				}

				encoded := base64.StdEncoding.EncodeToString(buf)
				progress <- fmt.Sprintf("data:image/png;base64,%s", encoded)
			},
		},
	}
}
