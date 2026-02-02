package devbrowser

import (
	"fmt"

	"github.com/chromedp/chromedp"
)

func (b *DevBrowser) getManagementTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_set_viewport",
			Description: "Set browser viewport size to desktop, mobile, or tablet presets. This change is persisted for future sessions.",
			Parameters: []ParameterMetadata{
				{
					Name:        "mode",
					Description: "Viewport mode: 'desktop' (1440x900), 'mobile' (375x812), or 'tablet' (768x1024)",
					Required:    true,
					Type:        "string",
					EnumValues:  []string{"desktop", "mobile", "tablet"},
				},
			},
			Execute: func(args map[string]any) {
				mode, ok := args["mode"].(string)
				if !ok {
					b.Logger("Mode parameter is required")
					return
				}

				var w, h int
				switch mode {
				case "desktop":
					w, h = 1440, 900
				case "mobile":
					w, h = 375, 812
				case "tablet":
					w, h = 768, 1024
				default:
					b.Logger(fmt.Sprintf("Unknown mode: %s", mode))
					return
				}

				b.mu.Lock()
				b.width = w
				b.height = h
				b.mu.Unlock()

				if err := b.SaveConfig(); err != nil {
					b.Logger(fmt.Sprintf("Error saving viewport config: %v", err))
				}

				if b.isOpen && b.ctx != nil {
					err := chromedp.Run(b.ctx, chromedp.EmulateViewport(int64(w), int64(h)))
					if err != nil {
						b.Logger(fmt.Sprintf("Error applying viewport: %v", err))
						return
					}
					b.ui.RefreshUI()
				}

				b.Logger(fmt.Sprintf("Viewport set to %s (%dx%d)", mode, w, h))
			},
		},
	}
}
