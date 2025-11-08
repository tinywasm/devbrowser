# Browser Click Element Implementation

## Overview
Implement `browser_click_element` MCP tool to programmatically click DOM elements, enabling LLMs to test user interactions, trigger events, and verify interactive behaviors.

## Tool Specification

**Name**: `browser_click_element`

**Description**: "Click DOM element by CSS selector to test interactions, trigger events, or simulate user actions. Useful for testing buttons, links, and interactive components."

**Parameters**:
```go
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
}
```

## Implementation Requirements

### File Location
`devbrowser/mcp-interaction.go`

### Core Function
```go
func (b *DevBrowser) getInteractionTools() []ToolMetadata
```

### Dependencies
- chromedp.Click(selector string)
- chromedp.WaitVisible(selector string) - ensure element exists before click
- chromedp.Sleep() - for wait_after parameter

### Execute Logic
1. Check browser is open (`b.isOpen`)
2. Extract `selector` parameter (required)
3. Extract `wait_after` parameter (default: 100)
4. Validate selector is not empty
5. Wait for element to be visible: `chromedp.WaitVisible(selector, chromedp.ByQuery)`
6. Click element: `chromedp.Click(selector, chromedp.ByQuery)`
7. Sleep for wait_after milliseconds
8. Send success message via progress channel

### Error Handling
- Browser not open: "Browser is not open. Please open it first with browser_open"
- Empty selector: "Selector parameter is required"
- Element not found: "Element not found: {selector}"
- Element not visible: "Element not visible: {selector}"
- Element not clickable: "Element not clickable: {selector}"
- Timeout: "Timeout waiting for element: {selector}"

### Output Format
Success: `Clicked element: {selector}`

Error examples:
```
Element not found: #nonexistent-btn
Timeout waiting for element: .hidden-element
Element not clickable: #disabled-button
```

### Wait Strategy
Use chromedp's built-in waiting:
```go
chromedp.WaitVisible(selector, chromedp.ByQuery)
chromedp.Click(selector, chromedp.ByQuery)
```

Add configurable post-click wait for animations/effects:
```go
time.Sleep(time.Millisecond * time.Duration(wait_after))
```

## Testing Requirements

### File Location
`devbrowser/mcp-interaction_test.go`

### Test Cases

**TestClickButton**
- Navigate to page with button
- HTML: `<button id="test-btn">Click Me</button>`
- Add click handler that sets text
- Click selector: "#test-btn"
- Verify: success message
- Verify: handler executed (check DOM change)

**TestClickLink**
- Page with link: `<a href="#section" id="link">Go</a>`
- Click selector: "#link"
- Verify: success message
- Verify: navigation occurred (check URL fragment)

**TestClickWithClass**
- Button with class: `<button class="primary-btn">Submit</button>`
- Click selector: ".primary-btn"
- Verify: success message

**TestClickComplexSelector**
- Nested structure: `<div class="form"><button type="submit">Save</button></div>`
- Click selector: ".form button[type='submit']"
- Verify: success message

**TestClickElementNotFound**
- Click selector: "#nonexistent"
- Verify: error "Element not found"

**TestClickHiddenElement**
- Button with display:none
- Click selector: "#hidden-btn"
- Verify: error "Element not visible" or "Timeout"

**TestClickDisabledButton**
- Button with disabled attribute
- Click selector: "#disabled-btn"
- Attempt click
- Verify: either succeeds (chromedp allows) or error

**TestClickWaitAfter**
- Button that triggers animation (200ms)
- Click with wait_after=250
- Verify: animation completes before return

**TestClickMultipleElements**
- Click button 1
- Verify success
- Click button 2
- Verify success
- Check both actions executed

**TestClickEmptySelector**
- Click selector: ""
- Verify: error "Selector parameter is required"

**TestClickBrowserNotOpen**
- Don't open browser
- Click selector: "#test"
- Verify: error "Browser is not open"

### Integration Test
Add to `mcp-tools_integration_test.go`:
```go
func TestClickElementToolRegistered(t *testing.T) {
    db, _ := DefaultTestBrowser()
    tools := db.GetMCPToolsMetadata()
    
    found := false
    for _, tool := range tools {
        if tool.Name == "browser_click_element" {
            found = true
            // Verify has selector parameter (required)
            // Verify has wait_after parameter (optional)
        }
    }
    if !found {
        t.Error("browser_click_element not registered")
    }
}
```

## Implementation Notes

### Selector Types Supported
CSS selectors via chromedp.ByQuery:
- ID: `#submit-button`
- Class: `.nav-item`
- Attribute: `[data-testid="login"]`
- Tag: `button`
- Complex: `.form button[type="submit"]`
- Nth-child: `.items li:nth-child(2)`

### Click Mechanics
ChromeDP click behavior:
- Scrolls element into view
- Waits for element to be actionable
- Simulates real mouse click
- Triggers all associated events (mousedown, mouseup, click)

### Wait Considerations
`wait_after` parameter useful for:
- CSS animations
- JavaScript event handlers
- DOM updates
- Network requests triggered by click

Default 100ms covers most synchronous updates.

### Multiple Matches
If selector matches multiple elements, ChromeDP clicks the first:
- Document order
- Consider using more specific selectors
- Or use nth-child/nth-of-type

### Performance
- Element lookup: 10-50ms
- Click action: < 10ms
- wait_after: configurable
- Total: typically < 200ms

### Common Use Cases for LLM
```
// Test button click
browser_click_element(selector: "#submit-btn")

// Navigate via link
browser_click_element(selector: "a[href='/about']")

// Toggle dropdown
browser_click_element(selector: ".dropdown-toggle")

// Close modal
browser_click_element(selector: ".modal .close-btn")

// Submit form
browser_click_element(selector: "form button[type='submit']")
```

### Limitations
- Cannot right-click (contextmenu)
- Cannot double-click (use evaluate_js for dblclick event)
- Cannot drag-and-drop (too complex for this tool)
- Cannot hover (use evaluate_js to trigger mouseover)

### Alternative: Evaluate JS
For complex interactions, use `browser_evaluate_js`:
```javascript
// Double click
document.querySelector('#element').dispatchEvent(new MouseEvent('dblclick'))

// Right click
document.querySelector('#element').dispatchEvent(new MouseEvent('contextmenu'))

// Hover
document.querySelector('#element').dispatchEvent(new MouseEvent('mouseover'))
```

### Error Recovery
If click fails:
1. Check element exists: `browser_evaluate_js("document.querySelector('#id')")`
2. Check element visible: `browser_evaluate_js("getComputedStyle(document.querySelector('#id')).display")`
3. Check element enabled: `browser_evaluate_js("document.querySelector('#id').disabled")`
4. Use screenshot to verify page state

## Example Usage
```
LLM: "Click the submit button to test form submission"
Tool: browser_click_element(selector: "#submit-btn", wait_after: 200)
Output: Clicked element: #submit-btn

LLM: "Now check if the form was submitted"
Tool: browser_get_console()
Output: Form submitted successfully

---

LLM: "Try clicking the save button"
Tool: browser_click_element(selector: "#save")
Output: Element not found: #save

LLM: "The save button doesn't exist. Check HTML structure."
Tool: browser_evaluate_js(script: "document.body.innerHTML")
```

## API Reference
ChromeDP methods used:
- `chromedp.WaitVisible(selector string, opts ...QueryOption)` - Wait for element
- `chromedp.Click(selector string, opts ...QueryOption)` - Click element
- `chromedp.ByQuery` - Use CSS selector
- `chromedp.NodeVisible` - Check visibility
- `chromedp.NodeEnabled` - Check if enabled

Query options:
- `chromedp.ByQuery` - CSS selector (default)
- `chromedp.ByID` - ID only
- `chromedp.BySearch` - XPath or CSS

## Success Criteria
- [ ] Tool registers in GetMCPToolsMetadata()
- [ ] Clicks elements successfully
- [ ] Waits for elements to be visible
- [ ] Handles element not found gracefully
- [ ] wait_after parameter works
- [ ] Supports complex CSS selectors
- [ ] All tests pass
- [ ] Performance < 200ms typical case
