package devbrowser

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/input"
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
				{
					Name:        "timeout",
					Description: "Maximum milliseconds to wait for the element to be visible",
					Required:    false,
					Type:        "number",
					Default:     5000,
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

				timeoutMs := 5000.0
				if t, ok := args["timeout"].(float64); ok {
					timeoutMs = t
				}

				// Create context with timeout
				ctx, cancel := context.WithTimeout(b.ctx, time.Duration(timeoutMs)*time.Millisecond)
				defer cancel()

				// 1. Wait for element to be present in DOM (WaitReady)
				err := chromedp.Run(ctx, chromedp.WaitReady(selector, chromedp.ByQuery))
				if err != nil {
					if err == context.DeadlineExceeded {
						b.Logger(fmt.Sprintf("Timeout exceeded waiting for element presence: %s", selector))
					} else {
						b.Logger(fmt.Sprintf("Error waiting for element %s: %v", selector, err))
					}
					return
				}

				// 2. Attempt standard click with a short timeout (500ms)
				// If it fails (e.g., covered, not visible), fallback to JS click
				clickCtx, clickCancel := context.WithTimeout(ctx, 500*time.Millisecond)
				err = chromedp.Run(clickCtx, chromedp.Click(selector, chromedp.ByQuery))
				clickCancel()

				if err == nil {
					b.Logger(fmt.Sprintf("Clicked element: %s", selector))
				} else {
					// 3. Fallback: JavaScript click
					b.Logger(fmt.Sprintf("Standard click failed (%v), attempting JS fallback for: %s", err, selector))

					// Use strconv.Quote to safely escape the selector for JS string
					jsClick := fmt.Sprintf("document.querySelector(%s).click()", strconv.Quote(selector))

					if err := chromedp.Run(ctx, chromedp.Evaluate(jsClick, nil)); err != nil {
						b.Logger(fmt.Sprintf("JS click fallback failed for %s: %v", selector, err))
						return
					}
					b.Logger(fmt.Sprintf("Clicked element (JS fallback): %s", selector))
				}

				// Wait after action
				if waitAfter > 0 {
					chromedp.Run(ctx, chromedp.Sleep(time.Duration(waitAfter)*time.Millisecond))
				}
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
				{
					Name:        "timeout",
					Description: "Maximum milliseconds to wait for the element to be visible",
					Required:    false,
					Type:        "number",
					Default:     5000,
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

				timeoutMs := 5000.0
				if t, ok := args["timeout"].(float64); ok {
					timeoutMs = t
				}

				// Create context with timeout
				ctx, cancel := context.WithTimeout(b.ctx, time.Duration(timeoutMs)*time.Millisecond)
				defer cancel()

				err := chromedp.Run(ctx,
					chromedp.WaitVisible(selector, chromedp.ByQuery),
					chromedp.SendKeys(selector, value, chromedp.ByQuery),
					chromedp.Sleep(time.Duration(waitAfter)*time.Millisecond),
				)

				if err != nil {
					if err == context.DeadlineExceeded {
						b.Logger(fmt.Sprintf("Timeout exceeded waiting for element: %s", selector))
					} else {
						b.Logger(fmt.Sprintf("Error filling element %s: %v", selector, err))
					}
					return
				}

				b.Logger(fmt.Sprintf("Filled element %s with '%s'", selector, value))
			},
		},
		{
			Name:        "browser_swipe_element",
			Description: "Simulate a swipe gesture on an element (up, down, left, right).",
			Parameters: []ParameterMetadata{
				{
					Name:        "selector",
					Description: "CSS selector for the element to swipe on",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "direction",
					Description: "Direction of swipe: 'up', 'down', 'left', 'right'",
					Required:    true,
					Type:        "string",
					EnumValues:  []string{"up", "down", "left", "right"},
				},
				{
					Name:        "distance",
					Description: "Distance in pixels to swipe",
					Required:    true,
					Type:        "number",
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

				direction, ok := args["direction"].(string)
				if !ok || direction == "" {
					b.Logger("Direction parameter is required")
					return
				}

				distanceVal, ok := args["distance"].(float64)
				if !ok {
					b.Logger("Distance parameter is required")
					return
				}
				distance := int(distanceVal)

				// Create context with timeout
				ctx, cancel := context.WithTimeout(b.ctx, 5000*time.Millisecond)
				defer cancel()

				// Swipe Logic
				err := chromedp.Run(ctx,
					chromedp.WaitVisible(selector, chromedp.ByQuery),
					chromedp.ActionFunc(func(ctx context.Context) error {
						// 1. Get element dimensions to find center
						// We use javascript to get bounding client rect as it is reliable
						script := fmt.Sprintf(`
							(function() {
								const el = document.querySelector('%s');
								if (!el) return null;
								const rect = el.getBoundingClientRect();
								return {x: rect.left + rect.width/2, y: rect.top + rect.height/2};
							})()
						`, selector)

						var res map[string]float64
						if err := chromedp.Evaluate(script, &res).Do(ctx); err != nil {
							return err
						}

						startX := res["x"]
						startY := res["y"]
						endX := startX
						endY := startY

						switch direction {
						case "up":
							endY -= float64(distance)
						case "down":
							endY += float64(distance)
						case "left":
							endX -= float64(distance)
						case "right":
							endX += float64(distance)
						}

						// Perform Mouse sequence using cdproto/input
						// Move to start
						p1 := input.DispatchMouseEvent(input.MouseMoved, startX, startY)
						if err := p1.Do(ctx); err != nil {
							return err
						}
						// Mouse Down
						p2 := input.DispatchMouseEvent(input.MousePressed, startX, startY).WithButton("left").WithClickCount(1)
						if err := p2.Do(ctx); err != nil {
							return err
						}
						// Move to end (swipe)
						p3 := input.DispatchMouseEvent(input.MouseMoved, endX, endY).WithButton("left")
						if err := p3.Do(ctx); err != nil {
							return err
						}
						// Mouse Up
						p4 := input.DispatchMouseEvent(input.MouseReleased, endX, endY).WithButton("left").WithClickCount(1)
						if err := p4.Do(ctx); err != nil {
							return err
						}
						return nil
					}),
				)

				if err != nil {
					b.Logger(fmt.Sprintf("Error swiping element %s: %v", selector, err))
					return
				}

				b.Logger(fmt.Sprintf("Swiped %s on %s by %dpx", direction, selector, distance))
			},
		},
	}
}
