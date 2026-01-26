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
					chromedp.Evaluate(`
						(() => {
							const getStructure = (el, depth = 0) => {
								if (depth > 4 || !el) return '';
								
								const tag = el.tagName.toLowerCase();
								const indent = '  '.repeat(depth);
								const style = window.getComputedStyle(el);
								
								// Skip invisible elements
								if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') return '';
								
								// Get direct text content
								const directText = Array.from(el.childNodes)
									.filter(n => n.nodeType === 3)
									.map(n => n.textContent.trim())
									.filter(t => t)
									.join(' ');
								
								let result = indent + '<' + tag;
								if (el.id) result += ' id="' + el.id + '"';
								if (el.className) result += ' class="' + el.className + '"';
								
								// Add critical visual styles (only if non-default)
								const styles = [];
								if (style.display !== 'block' && style.display !== 'inline') styles.push('display:' + style.display);
								if (style.position !== 'static') styles.push('position:' + style.position);
								if (style.backgroundColor !== 'rgba(0, 0, 0, 0)' && style.backgroundColor !== 'transparent') {
									styles.push('bg:' + style.backgroundColor);
								}
								if (style.color !== 'rgb(0, 0, 0)') styles.push('color:' + style.color);
								if (parseInt(style.fontSize) > 16) styles.push('font:' + style.fontSize);
								if (style.fontWeight === 'bold' || parseInt(style.fontWeight) >= 700) styles.push('bold');
								const w = parseInt(style.width);
								const h = parseInt(style.height);
								if (w > 50) styles.push('w:' + w);
								if (h > 50) styles.push('h:' + h);
								
								if (styles.length > 0) result += ' [' + styles.join(' ') + ']';
								result += '>';
								
								if (directText) result += ' ' + directText;
								result += '\n';
								
								// Recurse for children
								Array.from(el.children).forEach(child => {
									result += getStructure(child, depth + 1);
								});
								
								return result;
							};
							return getStructure(document.body);
						})()
					`, &htmlStructure),
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
