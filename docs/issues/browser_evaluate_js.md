# Browser Evaluate JS Implementation

## Overview
Implement `browser_evaluate_js` MCP tool to execute arbitrary JavaScript in browser context, enabling LLMs to inspect DOM, test WASM functions, manipulate state, and perform advanced debugging.

## Tool Specification

**Name**: `browser_evaluate_js`

**Description**: "Execute JavaScript code in browser context to inspect DOM, call WASM exports, test functions, or debug application state. Returns execution result or error."

**Parameters**:
```go
{
    Name:        "script",
    Description: "JavaScript code to execute in browser context",
    Required:    true,
    Type:        "string",
},
{
    Name:        "await_promise",
    Description: "Wait for Promise resolution if script returns Promise",
    Required:    false,
    Type:        "boolean",
    Default:     false,
}
```

## Implementation Requirements

### File Location
`devbrowser/mcp-evaluate-js.go`

### Core Function
```go
func (b *DevBrowser) getEvaluateJsTools() []ToolMetadata
```

### Dependencies
- chromedp.Evaluate(expression string, res interface{})
- JSON serialization for complex results
- Error handling for script exceptions

### Execute Logic
1. Check browser is open (`b.isOpen`)
2. Extract `script` parameter (required)
3. Extract `await_promise` parameter (default: false)
4. Execute script using chromedp.Evaluate:
   - If await_promise=false: Direct evaluation
   - If await_promise=true: Use chromedp.EvaluateAsDevTools with awaitPromise
5. Capture result and errors
6. Format output:
   - Success: Return stringified result
   - Error: Return error message with "Error: " prefix
7. Send via progress channel

### Error Handling
- Browser not open: "Browser is not open. Please open it first with browser_open"
- Empty script: "Script parameter is required"
- Script exception: "Error: {exception message}"
- Evaluation timeout: "Script execution timeout"

### Result Formatting
- Primitives: Convert to string directly
- Objects/Arrays: JSON.stringify
- undefined: "undefined"
- Functions: "[Function]"
- Errors: "Error: {message}"

### Output Format
```
// Success examples:
42
"hello world"
{"foo": "bar", "count": 5}
[1, 2, 3]
true

// Error example:
Error: Cannot read property 'foo' of undefined
```

## Testing Requirements

### File Location
`devbrowser/mcp-evaluate-js_test.go`

### Test Cases

**TestEvaluateJsPrimitive**
- Open browser to blank page
- Evaluate: "2 + 2"
- Verify: output "4"

**TestEvaluateJsString**
- Evaluate: "'hello' + ' world'"
- Verify: output "hello world"

**TestEvaluateJsObject**
- Evaluate: "({name: 'test', value: 42})"
- Verify: JSON output contains name and value

**TestEvaluateJsArray**
- Evaluate: "[1, 2, 3].map(x => x * 2)"
- Verify: output "[2,4,6]"

**TestEvaluateJsDOMInspection**
- Navigate to page with <div id="test">Content</div>
- Evaluate: "document.getElementById('test').textContent"
- Verify: output "Content"

**TestEvaluateJsWASMAccess**
- Page with WASM module
- Evaluate: "typeof WebAssembly"
- Verify: output "object"

**TestEvaluateJsError**
- Evaluate: "throw new Error('test error')"
- Verify: output starts with "Error: test error"

**TestEvaluateJsUndefined**
- Evaluate: "undefined"
- Verify: output "undefined"

**TestEvaluateJsPromiseAwait**
- Evaluate: "Promise.resolve(42)" with await_promise=true
- Verify: output "42"

**TestEvaluateJsPromiseNoAwait**
- Evaluate: "Promise.resolve(42)" with await_promise=false
- Verify: output "[object Promise]" or similar

**TestEvaluateJsBrowserNotOpen**
- Don't open browser
- Evaluate: "1 + 1"
- Verify: error "Browser is not open"

**TestEvaluateJsEmptyScript**
- Open browser
- Evaluate: ""
- Verify: error "Script parameter is required"

### Integration Test
Add to `mcp-tools_integration_test.go`:
```go
func TestEvaluateJsToolRegistered(t *testing.T) {
    db, _ := DefaultTestBrowser()
    tools := db.GetMCPToolsMetadata()
    
    found := false
    for _, tool := range tools {
        if tool.Name == "browser_evaluate_js" {
            found = true
            // Verify has script parameter (required)
            // Verify has await_promise parameter (optional)
        }
    }
    if !found {
        t.Error("browser_evaluate_js not registered")
    }
}
```

## Implementation Notes

### Security Considerations
- Runs in same context as page (no sandbox)
- Can access all page globals
- Can modify DOM and state
- Trust assumption: LLM-generated scripts for development/debugging only

### Performance
- Simple expressions: < 10ms
- DOM queries: 10-50ms
- Complex computations: variable
- Promises with await: depends on resolution time

### Script Context
Scripts execute with access to:
- window object
- document (DOM)
- console
- WASM exports (if page loaded WASM)
- All page globals

### Common Use Cases for LLM
```javascript
// Inspect DOM element
document.querySelector('#myButton').disabled

// Check WASM module loaded
typeof Go !== 'undefined'

// Get computed styles
getComputedStyle(document.body).backgroundColor

// Test function
myGlobalFunction('test input')

// Modify for testing
document.getElementById('test').value = 'new value'

// Get app state
window.__appState || {}
```

### Limitations
- Cannot access Node.js APIs
- Cannot import ES modules directly
- Synchronous only (unless await_promise=true)
- Subject to page's CSP restrictions

### Error Categories
1. **Syntax errors**: Invalid JavaScript
2. **Runtime errors**: Exceptions during execution
3. **Timeout errors**: Long-running scripts
4. **Security errors**: CSP violations

## Example Usage
```
LLM: "Check if the submit button is disabled"
Tool: browser_evaluate_js(script="document.getElementById('submit').disabled")
Output: true

LLM: "Get all error messages on the page"
Tool: browser_evaluate_js(script="Array.from(document.querySelectorAll('.error')).map(e => e.textContent)")
Output: ["Invalid email", "Required field"]
```

## API Reference
ChromeDP methods used:
- `chromedp.Evaluate(expression string, res *interface{})` - basic evaluation
- `chromedp.EvaluateAsDevTools(expression string, res *interface{})` - with options
- Result types: string, number, boolean, object, array, null, undefined

## Success Criteria
- [ ] Tool registers in GetMCPToolsMetadata()
- [ ] Executes JavaScript successfully
- [ ] Returns correct result types (primitives, objects, arrays)
- [ ] Handles errors gracefully
- [ ] await_promise parameter works
- [ ] DOM access works
- [ ] All tests pass
- [ ] Performance < 100ms for simple scripts
