package devbrowser

import (
	"fmt"

	"github.com/chromedp/chromedp"
	"github.com/tinywasm/mcpserve"
)

func (b *DevBrowser) getManagementTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_emulate_device",
			Description: "Emulate a mobile device or tablet viewport without resizing the physical window. This toggle affects rendering and touch events. This change is persisted.",
			Parameters: []ParameterMetadata{
				{
					Name:        "mode",
					Description: "Device mode: 'desktop' (no emulation), 'mobile' (375x812), 'tablet' (768x1024), or 'off' (clear all overrides)",
					Required:    true,
					Type:        "string",
					EnumValues:  []string{"desktop", "mobile", "tablet", "off"},
				},
				{
					Name:        "capture",
					Description: "If true, automatically returns a screenshot of the emulated viewport (efficiency boost: switch + see in one step)",
					Required:    false,
					Type:        "boolean",
					Default:     false,
				},
				{
					Name:        "selector",
					Description: "If provided with capture=true, snippets only this element instead of the whole viewport",
					Required:    false,
					Type:        "string",
				},
			},
			Execute: func(args map[string]any) {
				mode, ok := args["mode"].(string)
				if !ok {
					b.Logger("Mode parameter is required")
					return
				}

				capture := false
				if c, ok := args["capture"].(bool); ok {
					capture = c
				}

				selector := ""
				if s, ok := args["selector"].(string); ok {
					selector = s
				}

				b.mu.Lock()
				b.viewportMode = mode
				b.mu.Unlock()

				if err := b.SaveConfig(); err != nil {
					b.Logger(fmt.Sprintf("Error saving emulation config: %v", err))
				}

				if b.isOpen && b.ctx != nil {
					if err := b.applyDeviceEmulation(); err != nil {
						b.Logger(fmt.Sprintf("Error applying emulation: %v", err))
						return
					}
					b.ui.RefreshUI()
				}

				statusMsg := fmt.Sprintf("Device emulation set to %s", mode)

				if capture {
					var res *ScreenshotResult
					var err error
					if selector != "" {
						res, err = b.CaptureElementScreenshot(selector)
					} else {
						res, err = b.CaptureScreenshot(false)
					}

					if err != nil {
						b.Logger(fmt.Sprintf("%s. Failed to capture screenshot: %v", statusMsg, err))
						return
					}

					// Build visual context report
					contextReport := fmt.Sprintf(
						"%s\nURL: %s | Title: %s | Viewport: %dx%d\n\n%s",
						statusMsg,
						res.PageURL,
						res.PageTitle,
						res.Width, res.Height,
						res.HTMLStructure,
					)

					b.Logger(contextReport, mcpserve.BinaryData{MimeType: "image/png", Data: res.ImageData})
				} else {
					b.Logger(statusMsg)
				}
			},
		},
	}
}

// applyDeviceEmulation applies the current b.viewportMode using CDP emulation commands.
func (b *DevBrowser) applyDeviceEmulation() error {
	b.mu.Lock()
	mode := b.viewportMode
	b.mu.Unlock()

	var w, h int64
	var mobile bool

	switch mode {
	case "mobile":
		w, h = 375, 812
		mobile = true
	case "tablet":
		w, h = 768, 1024
		mobile = true
	case "desktop", "off", "":
		// Clear overrides by emulating a standard desktop viewport
		// We use 0,0 to reset device metrics override if supported,
		// but standard practice is to just set it to the window size.
		b.mu.Lock()
		w, h = int64(b.width), int64(b.height)
		b.mu.Unlock()
		mobile = false
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	// Apply emulation which mimics the Toggle Device Toolbar behavior
	// We use EmulateViewport with explicit mobile flag.
	if mobile {
		return chromedp.Run(b.ctx, chromedp.EmulateViewport(w, h, chromedp.EmulateMobile))
	}

	return chromedp.Run(b.ctx, chromedp.EmulateViewport(w, h))
}
