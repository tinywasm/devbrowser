package devbrowser

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func (b *DevBrowser) getInteractionTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_click_element",
			Description: "Click DOM element by CSS selector to test interactions, trigger events, or simulate user actions. Useful for testing buttons, links, and interactive components.",
			Parameters: []ParameterMetadata{
				{
					Name:        "selector",
					Description: "CSS selector for element to click (e.g., '#submit-btn', '.nav-item')",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "wait_after",
					Description: "Milliseconds to wait after click for effects to complete",
					Required:    false,
					Type:        "number",
					Default:     100,
				},
			},
			Execute: func(args map[string]any) {
				if !b.isOpen {
					b.Logger("Browser is not open. Please open it first with browser_open")
					return
				}

				selector, ok := args["selector"].(string)
				if !ok || selector == "" {
					b.Logger("Selector parameter is required")
					return
				}

				waitAfter := 100
				if w, ok := args["wait_after"].(float64); ok {
					waitAfter = int(w)
				}

				err := chromedp.Run(b.ctx,
					chromedp.WaitVisible(selector, chromedp.ByQuery),
					chromedp.Click(selector, chromedp.ByQuery),
					chromedp.Sleep(time.Duration(waitAfter)*time.Millisecond),
				)

				if err != nil {
					b.Logger(fmt.Sprintf("Error clicking element %s: %v", selector, err))
					return
				}

				b.Logger(fmt.Sprintf("Clicked element: %s", selector))
			},
		},
		{
			Name:        "browser_fill_element",
			Description: "Fill a form field (input, textarea) with text. Simulates typing.",
			Parameters: []ParameterMetadata{
				{
					Name:        "selector",
					Description: "CSS selector for the input element (e.g., '#username')",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "value",
					Description: "Text value to enter",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "wait_after",
					Description: "Milliseconds to wait after typing",
					Required:    false,
					Type:        "number",
					Default:     100,
				},
			},
			Execute: func(args map[string]any) {
				if !b.isOpen {
					b.Logger("Browser is not open. Please open it first with browser_open")
					return
				}

				selector, ok := args["selector"].(string)
				if !ok || selector == "" {
					b.Logger("Selector parameter is required")
					return
				}

				value, ok := args["value"].(string)
				if !ok {
					b.Logger("Value parameter is required")
					return
				}

				waitAfter := 100
				if w, ok := args["wait_after"].(float64); ok {
					waitAfter = int(w)
				}

				err := chromedp.Run(b.ctx,
					chromedp.WaitVisible(selector, chromedp.ByQuery),
					chromedp.SendKeys(selector, value, chromedp.ByQuery),
					chromedp.Sleep(time.Duration(waitAfter)*time.Millisecond),
				)

				if err != nil {
					b.Logger(fmt.Sprintf("Error filling element %s: %v", selector, err))
					return
				}

				b.Logger(fmt.Sprintf("Filled element %s with '%s'", selector, value))
			},
		},
	}
}
