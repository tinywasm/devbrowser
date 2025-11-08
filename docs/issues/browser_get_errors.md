# Browser Get Errors Implementation

## Overview
Implement `browser_get_errors` MCP tool to capture JavaScript runtime errors and exceptions, providing LLMs with focused error information for quick debugging without filtering console logs.

## Tool Specification

**Name**: `browser_get_errors`

**Description**: "Get JavaScript runtime errors and uncaught exceptions to quickly identify crashes, bugs, or WASM panics. Returns error messages with stack traces."

**Parameters**:
```go
{
    Name:        "limit",
    Description: "Maximum number of recent errors to return",
    Required:    false,
    Type:        "number",
    Default:     20,
}
```

## Implementation Requirements

### File Location
`devbrowser/mcp-errors.go`

### Core Function
```go
func (b *DevBrowser) getErrorTools() []ToolMetadata
```

### New DevBrowser Fields
Add to `devbrowser.go`:
```go
type DevBrowser struct {
    // ... existing fields
    jsErrors     []JSError
    errorsMutex  sync.Mutex
}

type JSError struct {
    Message    string
    Source     string // File/URL where error occurred
    LineNumber int
    ColumnNumber int
    StackTrace string
    Timestamp  time.Time
}
```

### Dependencies
- `github.com/chromedp/cdproto/runtime`
- Runtime events: EventExceptionThrown

### Initialization
Add to network/console capture or create `initializeErrorCapture()`:
```go
func (b *DevBrowser) initializeErrorCapture() error {
    b.jsErrors = []JSError{}
    
    chromedp.ListenTarget(b.ctx, func(ev interface{}) {
        switch ev := ev.(type) {
        case *runtime.EventExceptionThrown:
            b.errorsMutex.Lock()
            defer b.errorsMutex.Unlock()
            
            exception := ev.ExceptionDetails
            jsErr := JSError{
                Message:      exception.Text,
                Source:       exception.URL,
                LineNumber:   int(exception.LineNumber),
                ColumnNumber: int(exception.ColumnNumber),
                Timestamp:    time.Now(),
            }
            
            // Extract stack trace if available
            if exception.Exception != nil && exception.Exception.Description != "" {
                jsErr.StackTrace = exception.Exception.Description
            }
            
            b.jsErrors = append(b.jsErrors, jsErr)
        case *runtime.EventExecutionContextsCleared:
            // Clear errors on page reload/navigation
            b.errorsMutex.Lock()
            b.jsErrors = []JSError{}
            b.errorsMutex.Unlock()
        }
    })
    
    return chromedp.Run(b.ctx, runtime.Enable())
}
```

### Execute Logic
1. Check browser is open (`b.isOpen`)
2. Extract `limit` parameter (default: 20)
3. Lock errorsMutex
4. Get most recent N errors
5. Format each error with location and stack
6. Send via progress channel

### Error Handling
- Browser not open: "Browser is not open. Please open it first with browser_open"
- No errors: "No JavaScript errors captured"

### Output Format
Each error as multi-line block:
```
Error: {message}
  at {source}:{line}:{column}
  {stackTrace}
---
```

Example:
```
Error: Cannot read property 'foo' of undefined
  at http://localhost:3000/script.js:42:15
  at HTMLButtonElement.onclick (http://localhost:3000/:1:1)
---
Error: Uncaught ReferenceError: undefinedFunc is not defined
  at http://localhost:3000/main.js:128:5
---
```

### Auto-Clear on Navigation
Clear errors when page reloads (via EventExecutionContextsCleared):
```go
case *runtime.EventExecutionContextsCleared:
    b.errorsMutex.Lock()
    b.jsErrors = []JSError{}
    b.errorsMutex.Unlock()
```

## Testing Requirements

### File Location
`devbrowser/mcp-errors_test.go`

### Test Cases

**TestCaptureJavaScriptError**
- Open browser to blank page
- Initialize error capture
- Execute: `throw new Error('test error')`
- Get errors
- Verify: error message contains "test error"
- Verify: source/line/column present

**TestCaptureReferenceError**
- Execute: `undefinedVariable.foo`
- Get errors
- Verify: "ReferenceError" in message
- Verify: stack trace present

**TestCaptureTypeError**
- Execute: `null.toString()`
- Get errors
- Verify: "TypeError" in message

**TestCaptureMultipleErrors**
- Throw 5 different errors
- Get errors
- Verify: all 5 captured
- Verify: in chronological order

**TestErrorsLimit**
- Throw 50 errors
- Get errors with limit=10
- Verify: exactly 10 most recent returned

**TestErrorsClearOnReload**
- Throw error
- Verify error captured
- Reload page
- Get errors
- Verify: no errors (cleared)

**TestErrorsWithStackTrace**
- Create nested function calls that throw
- Get errors
- Verify: stack trace shows function chain

**TestNoErrorsWhenNone**
- Open browser
- Navigate to clean page
- Get errors
- Verify: "No JavaScript errors captured"

**TestErrorsBrowserNotOpen**
- Don't open browser
- Get errors
- Verify: error "Browser is not open"

### Integration Test
Add to `mcp-tools_integration_test.go`:
```go
func TestErrorsToolRegistered(t *testing.T) {
    db, _ := DefaultTestBrowser()
    tools := db.GetMCPToolsMetadata()
    
    found := false
    for _, tool := range tools {
        if tool.Name == "browser_get_errors" {
            found = true
            // Verify has limit parameter
            // Verify limit default is 20
        }
    }
    if !found {
        t.Error("browser_get_errors not registered")
    }
}
```

## Implementation Notes

### Error Types Captured
- Syntax errors (during script load)
- Runtime errors (TypeError, ReferenceError, etc.)
- Uncaught exceptions
- Promise rejections (may require separate handling)
- WASM panics (if they surface as JS errors)

### Stack Trace Parsing
ChromeDP provides stack traces in `exception.Exception.Description`:
- Parse to extract function names
- Include file URLs
- Show line/column numbers
- Format for readability

### Performance Considerations
- Error events are relatively rare (good performance)
- Mutex locking is lightweight
- Consider max buffer size (e.g., 1000 errors)

### Timestamp Utility
Include timestamp to show when error occurred:
```go
Timestamp: time.Now()
```

Format in output:
```
[2025-11-08 14:32:45] Error: message
```

### Promise Rejection Handling
May need additional listener for unhandled promise rejections:
```go
case *runtime.EventConsoleAPICalled:
    if ev.Type == "error" {
        // Also capture console.error calls
    }
```

### Common Use Cases for LLM
- "Check if there are any JavaScript errors"
- "Find the error causing the page to crash"
- "Debug WASM panic message"
- "Identify which function is throwing"
- "Check for uncaught exceptions"

### Error vs Console Log
Errors tool is focused subset of console logs:
- Only errors/exceptions
- Includes stack traces
- Shows source location
- Easier for LLM to parse

### Relation to Console Tool
`browser_get_errors` is complementary to `browser_get_console`:
- Console: All log levels (log, warn, error, info)
- Errors: Only runtime errors with stack traces

## Example Usage
```
LLM: "Check if there are any errors on the page"
Tool: browser_get_errors()
Output:
Error: Cannot read property 'value' of null
  at http://localhost:3000/app.js:156:32
  at HTMLButtonElement.onclick (http://localhost:3000/:1:1)
---

LLM: "The app.js file is trying to access a null element at line 156"
```

## API Reference
ChromeDP runtime events:
- `runtime.EventExceptionThrown` - JavaScript exception occurred
- `runtime.EventExceptionDetails` - Details: message, source, line, stack
- `runtime.Enable()` - Start capturing runtime events

Exception structure:
```go
type ExceptionDetails struct {
    Text        string          // Error message
    LineNumber  int64           // Line where error occurred
    ColumnNumber int64          // Column number
    ScriptID    string          // Script identifier
    URL         string          // Script URL/filename
    Exception   *RemoteObject   // Exception object with stack
}
```

## Success Criteria
- [ ] Tool registers in GetMCPToolsMetadata()
- [ ] Captures JavaScript errors
- [ ] Records stack traces
- [ ] Shows source location (file, line, column)
- [ ] Auto-clears on navigation
- [ ] Limit parameter works
- [ ] All tests pass
- [ ] Performance < 10ms per error event
