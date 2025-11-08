# Browser Screenshot Implementation

## Overview
Implement `browser_screenshot` MCP tool to capture current browser state as PNG image, enabling LLMs to verify visual rendering, detect layout issues, and confirm UI changes.

## Tool Specification

**Name**: `browser_screenshot`

**Description**: "Capture screenshot of current browser viewport to verify visual rendering, layout correctness, or UI state. Returns base64-encoded PNG image."

**Parameters**:
```go
{
    Name:        "fullpage",
    Description: "Capture full page height instead of viewport only",
    Required:    false,
    Type:        "boolean",
    Default:     false,
}
```

## Implementation Requirements

### File Location
`devbrowser/mcp-screenshot.go`

### Core Function
```go
func (b *DevBrowser) getScreenshotTools() []ToolMetadata
```

### Dependencies
- chromedp.CaptureScreenshot (viewport)
- chromedp.FullScreenshot (full page)
- base64 encoding for output

### Execute Logic
1. Check browser is open (`b.isOpen`)
2. Extract `fullpage` parameter from args (default: false)
3. Use chromedp to capture screenshot:
   - If fullpage=false: `chromedp.CaptureScreenshot(&buf)`
   - If fullpage=true: `chromedp.FullScreenshot(&buf, 90)` (quality 90)
4. Encode buffer to base64
5. Send via progress channel with format: `data:image/png;base64,{encoded}`

### Error Handling
- Browser not open: "Browser is not open. Please open it first with browser_open"
- Capture fails: "Failed to capture screenshot: {error}"
- Empty buffer: "Screenshot capture returned empty buffer"

### Output Format
Send single message via progress channel:
```
data:image/png;base64,iVBORw0KGgoAAAANSUhEUg...
```

LLM can interpret base64 data URI directly in most contexts.

## Testing Requirements

### File Location
`devbrowser/mcp-screenshot_test.go`

### Test Cases

**TestScreenshotViewport**
- Open browser to blank page
- Initialize screenshot tool
- Capture viewport screenshot
- Verify: base64 output starts with "data:image/png;base64,"
- Verify: decoded PNG has valid header (89 50 4E 47)
- Verify: buffer size > 0

**TestScreenshotFullPage**
- Open browser to page with content exceeding viewport
- Add div with height > 2000px
- Capture with fullpage=true
- Verify: base64 output present
- Verify: image height > viewport height

**TestScreenshotBrowserNotOpen**
- Create DevBrowser instance without opening
- Call screenshot tool
- Verify: error message "Browser is not open"

**TestScreenshotAfterPageLoad**
- Open browser
- Navigate to page with specific content
- Wait for render
- Capture screenshot
- Verify: successful capture

### Integration Test
Add to `mcp-tools_integration_test.go`:
```go
func TestScreenshotToolRegistered(t *testing.T) {
    db, _ := DefaultTestBrowser()
    tools := db.GetMCPToolsMetadata()
    
    found := false
    for _, tool := range tools {
        if tool.Name == "browser_screenshot" {
            found = true
            // Verify has fullpage parameter
            // Verify parameter defaults
        }
    }
    if !found {
        t.Error("browser_screenshot not registered")
    }
}
```

## Implementation Notes

### Performance Considerations
- Viewport screenshots are fast (~50-100ms)
- Full page screenshots can take 200-500ms for long pages
- Base64 encoding adds ~33% to size
- Consider adding timeout for very long pages

### Size Optimization
- Use PNG compression (already default in chromedp)
- Quality parameter 90 balances size/fidelity
- Viewport captures are smaller than fullpage

### Alternative Formats
Future enhancement: Add format parameter (png/jpeg) if size becomes issue

### Use Cases for LLM
- Verify CSS changes applied correctly
- Detect layout breakage after code changes
- Confirm responsive design at different viewports
- Visual regression testing
- Debugging WASM rendering issues

## Example Usage
```
LLM: "Take a screenshot to verify the button color changed"
Tool: browser_screenshot(fullpage=false)
Output: data:image/png;base64,iVBORw0K...
LLM: [analyzes image] "Confirmed, button is now blue"
```

## API Reference
ChromeDP methods used:
- `chromedp.CaptureScreenshot(res *[]byte)` - viewport only
- `chromedp.FullScreenshot(res *[]byte, quality int)` - full page
- `chromedp.EmulateViewport(width, height int64)` - if viewport control needed

## Success Criteria
- [ ] Tool registers in GetMCPToolsMetadata()
- [ ] Viewport screenshot captures successfully
- [ ] Full page screenshot works for long pages
- [ ] Base64 output is valid PNG data URI
- [ ] Error handling for browser not open
- [ ] All tests pass
- [ ] Performance < 500ms for typical pages
