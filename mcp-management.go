package devbrowser

import (
	"fmt"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/tinywasm/mcpserve"
)

func (b *DevBrowser) getManagementTools() []mcpserve.ToolMetadata {
	return []mcpserve.ToolMetadata{
		{
			Name:        "browser_emulate_device",
			Description: "Emulate a mobile device or tablet viewport without resizing the physical window. This toggle affects rendering and touch events. This change is persisted.",
			Parameters: []mcpserve.ParameterMetadata{
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

	var actions []chromedp.Action

	switch mode {
	case "mobile":
		actions = append(actions,
			chromedp.EmulateViewport(375, 812, chromedp.EmulateMobile),
			emulation.SetTouchEmulationEnabled(true),
		)
	case "tablet":
		actions = append(actions,
			chromedp.EmulateViewport(768, 1024, chromedp.EmulateMobile),
			emulation.SetTouchEmulationEnabled(true),
		)
	case "desktop", "off", "":
		// Clear overrides by emulating a standard desktop viewport
		// We use ClearDeviceMetricsOverride to reset to the window size,
		// allowing the browser layout to adjust naturally to DevTools.
		actions = append(actions,
			emulation.ClearDeviceMetricsOverride(),
			emulation.SetTouchEmulationEnabled(false),
		)
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	return chromedp.Run(b.ctx, actions...)
}
