# Browser Network Logs Implementation

## Overview
Implement `browser_get_network_logs` MCP tool to capture network requests/responses, enabling LLMs to debug API failures, CORS issues, 404 errors, and asset loading problems.

## Tool Specification

**Name**: `browser_get_network_logs`

**Description**: "Get network requests and responses to debug API calls, asset loading failures, CORS errors, or slow requests. Shows URL, status, method, and timing."

**Parameters**:
```go
{
    Name:        "filter",
    Description: "Filter by request type",
    Required:    false,
    Type:        "string",
    EnumValues:  []string{"all", "xhr", "fetch", "document", "script", "image"},
    Default:     "all",
},
{
    Name:        "limit",
    Description: "Maximum number of recent requests to return",
    Required:    false,
    Type:        "number",
    Default:     50,
}
```

## Implementation Requirements

### File Location
`devbrowser/mcp-network.go`

### Core Function
```go
func (b *DevBrowser) getNetworkTools() []ToolMetadata
```

### New DevBrowser Fields
Add to `devbrowser.go`:
```go
type DevBrowser struct {
    // ... existing fields
    networkLogs []NetworkLogEntry
    networkMutex sync.Mutex
}

type NetworkLogEntry struct {
    URL        string
    Method     string
    Status     int
    Type       string // xhr, fetch, document, script, image, etc.
    Duration   int64  // milliseconds
    Failed     bool
    ErrorText  string
}
```

### Dependencies
- `github.com/chromedp/cdproto/network`
- Network domain events: RequestWillBeSent, ResponseReceived, LoadingFailed

### Initialization
Add to `initializeConsoleCapture()` or create `initializeNetworkCapture()`:
```go
func (b *DevBrowser) initializeNetworkCapture() error {
    b.networkLogs = []NetworkLogEntry{}
    
    chromedp.ListenTarget(b.ctx, func(ev interface{}) {
        switch ev := ev.(type) {
        case *network.EventRequestWillBeSent:
            // Track request start
        case *network.EventResponseReceived:
            // Record response details
        case *network.EventLoadingFailed:
            // Mark request as failed
        }
    })
    
    return chromedp.Run(b.ctx, network.Enable())
}
```

### Execute Logic
1. Check browser is open (`b.isOpen`)
2. Extract `filter` parameter (default: "all")
3. Extract `limit` parameter (default: 50)
4. Lock networkMutex
5. Filter logs by type if filter != "all"
6. Limit to most recent N entries
7. Format each entry as single line
8. Send via progress channel

### Error Handling
- Browser not open: "Browser is not open. Please open it first with browser_open"
- No logs: "No network requests captured"
- Empty after filter: "No {filter} requests found"

### Output Format
Each line: `{status} {method} {url} ({duration}ms) [type]`

Examples:
```
200 GET http://localhost:3000/main.wasm (245ms) [document]
200 GET http://localhost:3000/style.css (12ms) [stylesheet]
404 GET http://localhost:3000/missing.js (8ms) [script]
200 POST http://localhost:3000/api/users (156ms) [xhr]
Failed GET http://localhost:3000/timeout (30000ms) [fetch] - net::ERR_TIMED_OUT
```

### Auto-Clear on Navigation
Similar to console logs, clear network logs when page reloads:
```go
case *network.EventRequestWillBeSent:
    if ev.Type == "Document" {
        b.networkMutex.Lock()
        b.networkLogs = []NetworkLogEntry{}
        b.networkMutex.Unlock()
    }
```

## Testing Requirements

### File Location
`devbrowser/mcp-network_test.go`

### Test Cases

**TestNetworkLogsCapture**
- Open browser
- Initialize network capture
- Navigate to page that loads assets
- Get network logs
- Verify: multiple entries present
- Verify: includes document request

**TestNetworkLogsXHRFilter**
- Page makes XHR request
- Execute: fetch('/api/test')
- Get logs with filter="xhr"
- Verify: only XHR requests returned

**TestNetworkLogsFetchFilter**
- Page makes fetch() call
- Get logs with filter="fetch"
- Verify: fetch requests returned

**TestNetworkLogs404**
- Request non-existent resource
- Get network logs
- Verify: 404 status captured
- Verify: URL matches failed request

**TestNetworkLogsCORS**
- Make cross-origin request
- Get network logs
- Verify: CORS error captured (Failed=true)

**TestNetworkLogsLimit**
- Make 100 requests
- Get logs with limit=10
- Verify: exactly 10 most recent returned

**TestNetworkLogsClearOnReload**
- Make requests
- Verify logs present
- Reload page
- Verify logs cleared
- Make new request
- Verify only new request in logs

**TestNetworkLogsBrowserNotOpen**
- Don't open browser
- Get network logs
- Verify: error "Browser is not open"

### Integration Test
Add to `mcp-tools_integration_test.go`:
```go
func TestNetworkLogsToolRegistered(t *testing.T) {
    db, _ := DefaultTestBrowser()
    tools := db.GetMCPToolsMetadata()
    
    found := false
    for _, tool := range tools {
        if tool.Name == "browser_get_network_logs" {
            found = true
            // Verify has filter parameter
            // Verify has limit parameter
            // Verify filter enum values
        }
    }
    if !found {
        t.Error("browser_get_network_logs not registered")
    }
}
```

## Implementation Notes

### Request Tracking
Use request ID to correlate events:
- RequestWillBeSent provides ID, URL, method, type
- ResponseReceived provides status code
- LoadingFailed provides error

Store temporary map: `requestID -> NetworkLogEntry` during capture, then append to logs on completion.

### Performance Considerations
- Network events are frequent (dozens per page load)
- Mutex locking must be efficient
- Consider circular buffer if logs grow large
- Clear on navigation to prevent memory issues

### Type Classification
ChromeDP network types:
- Document: HTML pages
- Stylesheet: CSS files
- Script: JavaScript files
- Image: PNG, JPEG, etc.
- Font: WOFF, TTF
- XHR: XMLHttpRequest
- Fetch: fetch() API
- WebSocket: WS connections

### Timing Calculation
Store start timestamp on RequestWillBeSent, calculate duration on ResponseReceived:
```go
duration := responseTime - requestTime
```

### Common Use Cases for LLM
- "Check if API returned 500 error"
- "Verify WASM file loaded successfully"
- "Find which requests are slow (>1s)"
- "Check for 404 errors"
- "Debug CORS issues"
- "Verify all assets loaded"

### CORS Detection
Failed requests with specific error text:
- "net::ERR_BLOCKED_BY_CLIENT"
- "net::ERR_CORS_"
- "net::ERR_FAILED" (generic)

## Example Usage
```
LLM: "Check if the API call failed"
Tool: browser_get_network_logs(filter="xhr")
Output:
500 POST http://localhost:3000/api/users (234ms) [xhr]
200 GET http://localhost:3000/api/config (45ms) [xhr]

LLM: "The POST to /api/users returned 500"
```

## API Reference
ChromeDP network events:
- `network.EventRequestWillBeSent` - Request initiated
- `network.EventResponseReceived` - Response headers received
- `network.EventLoadingFinished` - Response complete
- `network.EventLoadingFailed` - Request failed
- `network.Enable()` - Start capturing

## Success Criteria
- [ ] Tool registers in GetMCPToolsMetadata()
- [ ] Captures all network requests
- [ ] Records status codes correctly
- [ ] Filters by type work
- [ ] Failed requests marked appropriately
- [ ] Auto-clears on navigation
- [ ] All tests pass
- [ ] Performance acceptable (< 1ms per event)
