# PLAN: Add GetLog() to DevBrowser

## Related Docs
- Depends on: [tinywasm/mcp PLAN.md](../../mcp/docs/PLAN.md) — must be completed first.

---

## Development Rules
- No external libraries; standard library only.
- Use `gotest` to run tests. Use `gopush` to publish.

---

## Context

After `tinywasm/mcp` adds `GetLog()` to the `Loggable` interface, `DevBrowser` must implement it. This is the **only change needed** — no tool `Execute` closures need modification.

---

## Step 1 — Add `GetLog()` method

**File:** `devbrowser.go`

```go
func (b *DevBrowser) GetLog() func(message ...any) {
    return b.log
}
```

---

## Publish
After tests pass: `gopush 'feat: implement GetLog for MCP Loggable interface'`
