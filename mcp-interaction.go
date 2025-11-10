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
			Execute: func(args map[string]any, progress chan<- any) {
				if !b.isOpen {
					progress <- "Browser is not open. Please open it first with browser_open"
					return
				}

				selector, ok := args["selector"].(string)
				if !ok || selector == "" {
					progress <- "Selector parameter is required"
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
					progress <- fmt.Sprintf("Error clicking element %s: %v", selector, err)
					return
				}

				progress <- fmt.Sprintf("Clicked element: %s", selector)
			},
		},
	}
}
