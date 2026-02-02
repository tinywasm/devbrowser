package devbrowser

import (
	"fmt"

	"github.com/chromedp/chromedp"
)

// getStructureJS is the JavaScript used to extract the page structure for LLM understanding.
const getStructureJS = `
(() => {
	const getStructure = (el, depth = 0) => {
		if (depth > 12 || !el) return ''; // Increased depth for more detail
		
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
		
		// Handle class names more cleanly
		if (el.className && typeof el.className === 'string') {
			const classes = el.className.split(/\s+/).filter(c => c).join(' ');
			if (classes) result += ' class="' + classes + '"';
		}
		
		// Add ARIA roles and labels which are critical for LLMs
		const role = el.getAttribute('role');
		const ariaLabel = el.getAttribute('aria-label');
		const placeholder = el.getAttribute('placeholder');
		const value = el.value;
		const name = el.getAttribute('name');
		const type = el.getAttribute('type');

		if (role) result += ' role="' + role + '"';
		if (ariaLabel) result += ' aria-label="' + ariaLabel + '"';
		if (placeholder) result += ' placeholder="' + placeholder + '"';
		if (name) result += ' name="' + name + '"';
		if (type) result += ' type="' + type + '"';
		if (value && tag !== 'body') result += ' value="' + value + '"';

		// Add critical visual styles (only if non-default)
		const styles = [];
		if (style.display !== 'block' && style.display !== 'inline' && style.display !== 'inline-block') {
			styles.push('display:' + style.display);
		}
		if (style.position !== 'static') styles.push('position:' + style.position);
		
		// Indicate interactive items
		const isClickable = window.getComputedStyle(el).cursor === 'pointer' || 
						   tag === 'button' || tag === 'a' || tag === 'input' || tag === 'select';
		if (isClickable) styles.push('clickable');

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
`

func (b *DevBrowser) getStructureTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_get_content",
			Description: "Get a text-based representation of the page content, optimized for LLM reading. Reduced token count compared to screenshots.",
			Parameters:  []ParameterMetadata{},
			Execute: func(args map[string]any) {
				if !b.isOpen {
					b.Logger("Browser is not open. Please open it first with browser_open")
					return
				}

				var pageTitle, pageURL, structure string
				var windowWidth, windowHeight int

				err := chromedp.Run(b.ctx,
					chromedp.Title(&pageTitle),
					chromedp.Location(&pageURL),
					chromedp.Evaluate(`window.innerWidth`, &windowWidth),
					chromedp.Evaluate(`window.innerHeight`, &windowHeight),
					chromedp.Evaluate(getStructureJS, &structure),
				)

				if err != nil {
					b.Logger(fmt.Sprintf("Failed to get page structure: %v", err))
					return
				}

				report := fmt.Sprintf(
					"URL: %s\nTitle: %s\nViewport: %dx%d\n\n%s",
					pageURL,
					pageTitle,
					windowWidth, windowHeight,
					structure,
				)

				b.Logger(report)
			},
		},
	}
}
