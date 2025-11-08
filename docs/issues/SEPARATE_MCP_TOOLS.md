# MCP Tools Separation - Refactoring Plan

## Problem
Current `mcp-tools.go` contains all MCP tool definitions in a single file with a monolithic `GetMCPToolsMetadata()` function. This creates maintainability issues:
- Hard to locate specific tool implementations
- File grows large as new tools are added
- Difficult to test individual tools in isolation
- No clear separation of concerns

## Solution
Separate MCP tools into focused, single-responsibility files organized by domain:

### File Structure
```
devbrowser/
├── mcp-tools.go           # Core types (ToolExecutor, ToolMetadata, ParameterMetadata)
├── mcp-management.go      # browser_open, browser_close, browser_reload
├── mcp-console.go         # browser_get_console
├── mcp-screenshot.go      # browser_screenshot (NEW)
├── mcp-evaluate-js.go     # browser_evaluate_js (NEW)
├── mcp-network.go         # browser_get_network_logs (NEW)
├── mcp-errors.go          # browser_get_errors (NEW)
├── mcp-interaction.go     # browser_click_element (NEW)
```

## Refactoring Steps

### Phase 1: Extract Core Types (mcp-tools.go)
Keep only shared types and interfaces:
```go
// ToolExecutor, ToolMetadata, ParameterMetadata remain
// GetMCPToolsMetadata() becomes aggregator function
```

New `GetMCPToolsMetadata()` signature:
```go
func (b *DevBrowser) GetMCPToolsMetadata() []ToolMetadata {
    tools := []ToolMetadata{}
    tools = append(tools, b.getManagementTools()...)
    tools = append(tools, b.getConsoleTools()...)
    tools = append(tools, b.getScreenshotTools()...)
    tools = append(tools, b.getEvaluateJsTools()...)
    tools = append(tools, b.getNetworkTools()...)
    tools = append(tools, b.getErrorTools()...)
    tools = append(tools, b.getInteractionTools()...)
    return tools
}
```

### Phase 2: Create Domain-Specific Files
Each file implements private method returning []ToolMetadata:

**mcp-management.go**
```go
func (b *DevBrowser) getManagementTools() []ToolMetadata {
    return []ToolMetadata{
        // browser_open, browser_close, browser_reload
    }
}
```

**mcp-console.go**
```go
func (b *DevBrowser) getConsoleTools() []ToolMetadata {
    return []ToolMetadata{
        // browser_get_console
    }
}
```

### Phase 3: Migrate Existing Tools
1. Move browser_open, browser_close, browser_reload to `mcp-management.go`
2. Move browser_get_console to `mcp-console.go`
3. Update tests to verify tools still register correctly
4. Ensure backward compatibility (same tool names/parameters)

### Phase 4: Implement New Tools
Follow implementation prompts (one tool per file):

- [browser_screenshot.md](browser_screenshot.md) - Screenshot capture implementation
- [browser_evaluate_js.md](browser_evaluate_js.md) - JavaScript evaluation implementation
- [browser_get_network_logs.md](browser_get_network_logs.md) - Network monitoring implementation
- [browser_get_errors.md](browser_get_errors.md) - Error tracking implementation
- [browser_click_element.md](browser_click_element.md) - Element interaction implementation

## Testing Strategy

### Important: Test Execution Constraint
**If browser-based tests cannot be executed in the current environment:**
- Proceed with implementation only (code creation)
- Skip test execution temporarily
- Mark tests as "pending review/execution"
- Implementation will be validated later when browser environment is available
- Do NOT block progress on test execution failures
- Focus on correct implementation structure and logic

### Unit Tests
Each domain file gets corresponding test file:
```
mcp-management_test.go
mcp-console_test.go
mcp-screenshot_test.go
mcp-evaluate-js_test.go
mcp-network_test.go
mcp-errors_test.go
mcp-interaction_test.go
```

**Test Creation Priority:**
1. Create test files with comprehensive test cases
2. Attempt test execution if environment permits
3. If tests fail due to environment (not code), document and skip
4. Tests will be validated in proper environment later

### Integration Test
Create `mcp-tools_integration_test.go` to verify:
- All tools register correctly via GetMCPToolsMetadata()
- Tool count matches expected
- No duplicate tool names
- All tools have valid metadata

Example:
```go
func TestGetMCPToolsMetadata_AllToolsRegistered(t *testing.T) {
    db, _ := DefaultTestBrowser()
    tools := db.GetMCPToolsMetadata()
    
    expectedToolNames := []string{
        "browser_open", "browser_close", "browser_reload",
        "browser_get_console", "browser_screenshot",
        "browser_evaluate_js", "browser_get_network_logs",
        "browser_get_errors", "browser_click_element",
    }
    
    // Verify all expected tools present
    // Verify no duplicates
    // Verify metadata completeness
}
```

## Migration Checklist
- [ ] Extract core types to mcp-tools.go
- [ ] Create mcp-management.go (move existing 3 tools)
- [ ] Create mcp-console.go (move existing 1 tool)
- [ ] Update GetMCPToolsMetadata() to aggregate
- [ ] Run existing tests (if environment permits, skip if not)
- [ ] Create integration test
- [ ] Implement browser_screenshot (mcp-screenshot.go + test file)
- [ ] Implement browser_evaluate_js (mcp-evaluate-js.go + test file)
- [ ] Implement browser_get_network_logs (mcp-network.go + test file)
- [ ] Implement browser_get_errors (mcp-errors.go + test file)
- [ ] Implement browser_click_element (mcp-interaction.go + test file)
- [ ] Update documentation
- [ ] Validate all tests in proper environment (deferred if needed)

## Benefits
- **Maintainability**: Each tool isolated in focused file
- **Testability**: Independent test files per domain
- **Scalability**: Easy to add new tools without touching existing
- **Clarity**: Clear separation of concerns
- **Collaboration**: Multiple developers can work on different tools

## Notes
- Maintain backward compatibility during migration
- Each new tool must include comprehensive tests
- Follow existing patterns (ToolMetadata structure, progress channel usage)
- Keep tool names consistent with existing conventions (browser_*)
