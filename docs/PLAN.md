# Plan: Migrate `devbrowser` to New `tinywasm/mcp` API

This plan outlines the steps to refactor all MCP-prefixed files in `tinywasm/devbrowser` to comply with the updated `tinywasm/mcp` API, utilizing `ormc` for automatic JSON Schema generation and validation, and reorganizing tests.

## 1. Prerequisites & Tooling
The new API relies on `ormc` for generating schemas and validation logic from Go structs.

- [ ] Install `ormc`:
  ```bash
  go install github.com/tinywasm/orm/cmd/ormc@latest
  ```

## 2. Model Centralization (`models.go`)
Create a new file `tinywasm/devbrowser/models.go` to store all tool argument structures. This allows a single `go generate` pass and keeps the codebase clean.

- [ ] Define argument structures in `models.go` (done).

## 3. Implementation Phases

### Phase 1: Core Tool Updates
Update files where the tools are simple and frequently used.

- [ ] **mcp-management.go**: Migrate `browser_emulate_device`.
- [ ] **mcp-console.go**: Migrate `browser_get_console`.
- [ ] **mcp-screenshot.go**: Migrate `browser_screenshot`.
- [ ] **mcp-navigation.go**: Migrate `browser_navigate`.
- [ ] **mcp-inspect.go**: Migrate `browser_inspect_element`.

### Phase 2: Interaction & Structure
Refactor more complex interaction tools and the structure extraction tool.

- [ ] **mcp-interaction.go**: Migrate `browser_click_element`, `browser_fill_element`, and `browser_swipe_element`.
- [ ] **mcp-structure.go**: Migrate `browser_get_content`.

### Phase 3: Advanced & Optional Tools
Migrate and re-enable tools that were previously commented out in `mcp-tools.go`.

- [ ] **mcp-performance.go**: Migrate performance analysis tools.
- [ ] **mcp-network.go**: Migrate network monitoring tools.
- [ ] **mcp-evaluate-js.go**: Migrate JS evaluation tools.
- [ ] **mcp-errors.go**: Migrate error reporting tools.

## 4. Test Reorganization
Organize all existing tests into a dedicated `tests/` directory to clean up the package root.

- [ ] Create `tinywasm/devbrowser/tests/` directory.
- [ ] Move all `*_test.go` files from root to `tests/`.
- [ ] Update test package names if necessary (typically to `devbrowser_test`).
- [ ] Fix imports in tests to point back to the parent `devbrowser` package.

**Files to move:**
- `console_capture_test.go`
- `console_logs_test.go`
- `context_flag_test.go`
- `default_test.go`
- `devbrowser_test.go`
- `device_emulation_test.go`
- `geometry_constraints_test.go`
- `geometry_test.go`
- `interaction_test.go`
- `logic_preservation_test.go`
- `mcp-performance_test.go`
- `mcp-tools_integration_test.go`
- `navigation_test.go`
- `robust_interaction_test.go`
- `structure_test.go`
- `swipe_test.go`

## 5. API Refactoring Details
1.  **Replace `Parameters`**: Remove `Parameters: []mcp.Parameter{...}` and replace with `InputSchema: new(XArgs).Schema()`.
2.  **Update `Execute` Signature**:
    - Old: `Execute: func(args map[string]any)`
    - New: `Execute: func(ctx *context.Context, req mcp.Request) (*mcp.Result, error)`
3.  **Use `req.Bind`**: Replace manual map extraction with `req.Bind(&args)`.
4.  **Use `mcp.Result`**: Return results using `mcp.Text()`, `mcp.JSON()`, or binary data where applicable.

## 6. Verification
- [ ] Run `ormc` in root project.
- [ ] Ensure all `mcp-*` files compile correctly.
- [ ] Update `mcp-tools.go` to uncomment and include all migrated tools.
- [ ] Run all tests from the new location: `go test ./tests/...`
