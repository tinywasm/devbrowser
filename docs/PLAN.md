# PLAN: fix broken chrome executable resolution + become wasmbrowsertest's chromedp backend

## Context / bug being fixed

`devbrowser` (`github.com/tinywasm/devbrowser`, at
`/home/cesar/Dev/Project/tinywasm/devbrowser`) vendors its own full copy of
`chromedp` under `devbrowser/chromedp/` (`allocate.go`, `conn.go`, `util.go`,
etc.) instead of importing the upstream module. That vendored copy contains
`findExecPath()`, which resolves the Chrome/Chromium binary to launch by
walking a fixed candidate list and returning the first one found via
`exec.LookPath`:

```go
locations = []string{
    "headless_shell", "headless-shell",
    "chromium", "chromium-browser",     // checked before google-chrome
    "google-chrome", "google-chrome-stable", ...
}
```

### Confirmed root cause

On this machine (Debian 13 "trixie"), two Chrome-family binaries are
installed:

- `/usr/bin/google-chrome` — launches fine in headless mode.
- `/usr/bin/chromium` (Debian package `chromium` 150.0.7871.46-1~deb13u1) —
  **crashes instantly** with `Trace/breakpoint trap` (a SIGTRAP) whenever
  launched with `--headless` or `--headless=new`, with or without
  `--no-sandbox`. Reproduced directly from a shell with no Go/chromedp code
  involved at all:

  ```
  $ chromium --headless --no-sandbox --user-data-dir=/tmp/x \
      --remote-debugging-port=0 about:blank
  Trace/breakpoint trap
  ```

  This is a packaging/environment regression in Debian's `chromium` build on
  this host — not a bug in devbrowser, chromedp, or devflow.

Because `chromium` is on `PATH` and is checked first, `findExecPath()`
always resolves to the broken binary instead of the working
`google-chrome`, so any devbrowser code path that lets chromedp
auto-resolve the executable (i.e. doesn't pass an explicit
`chromedp.ExecPath(...)`) silently launches a browser that dies before
printing anything, producing an unhelpful empty "chrome failed to start:"
error.

## Decision: devbrowser becomes the single owned chromedp integration

A sibling repo, `github.com/agnivade/wasmbrowsertest` (local fork at
`/home/cesar/Dev/Project/tinywasm/wasmbrowsertest`, used today by `gotest`
via `go test -exec wasmbrowsertest` to run WASM tests in a real browser),
imports `chromedp` directly and has the identical bug (see its own,
self-contained `docs/PLAN.md`). Maintaining two independent chromedp
integrations means bugs like this one get fixed twice and drift out of
sync. Since devbrowser is the more complete, actively-owned integration
(MCP tools, screenshots, console/network inspection, etc.), the direction
is: **fix the executable-resolution bug here, then expose an API that
wasmbrowsertest can call instead of importing chromedp itself.**
wasmbrowsertest keeps its current CLI contract (`-exec wasmbrowsertest`:
serve a wasm test binary over HTTP, drive it in a browser, relay the exit
code) but stops owning any chromedp code directly.

## Tasks

### Part A — fix the executable resolution bug (do this first, self-contained)

1. **Audit every allocator-construction call site** in this repo
   (`OpenBrowser.go`, `config.go`, `devbrowser.go`, the `mcp-*.go` files) for
   places where the Chrome executable path is left to chromedp's default
   resolution instead of being explicitly set via `chromedp.ExecPath(...)`.

2. **Add an exported `ResolveChromeExecPath()` helper** (new file, e.g.
   `execpath.go`) that:
   - Honors an explicit override first: an env var (e.g. `CHROME_EXECPATH`)
     and/or a field on whatever config struct `config.go` already defines.
   - Otherwise builds its own candidate list that checks
     `google-chrome`/`google-chrome-stable`/`google-chrome-beta` **before**
     `chromium`/`chromium-browser`/`headless_shell` — the inverse of
     upstream chromedp's default order — since Debian's `chromium` package
     has shown this crash-on-launch failure mode.
   - **Validates** each resolved candidate before accepting it: run
     `exec.Command(path, "--version")` with a short timeout (a couple of
     seconds); if it errors, times out, or the process dies to a signal,
     skip to the next candidate instead of handing a broken binary to
     chromedp. This makes resolution self-healing instead of a silent hard
     failure.
   - Returns a clear error naming every candidate tried and why each was
     rejected only if nothing usable is found at all.

3. **Wire `ResolveChromeExecPath()` into every call site found in Task 1**
   via `chromedp.ExecPath(path)`, replacing implicit default resolution.

4. **Add a regression test** (table-driven, following the existing style of
   `position_repro_test.go` in this repo) that fakes `PATH` with a small
   shell script named `chromium` that immediately exits/crashes, and another
   named `google-chrome` that prints a fake `DevTools listening on ...` line
   and stays alive; assert `ResolveChromeExecPath()` skips the crasher and
   returns the working one.

5. **Verify against the real repro**: with the fix in place, run devbrowser's
   own browser-launch path against this host (where the broken `chromium`
   150.0.7871.46-1~deb13u1 is still installed) and confirm it resolves to
   `google-chrome` and launches successfully.

### Part B — expose an API for wasmbrowsertest to consume (after Part A lands)

6. **Design and expose a browser-launch primitive** suited to
   wasmbrowsertest's needs: given a URL to navigate to, wait for an element
   (`#doneButton`) to become enabled, then evaluate a JS expression
   (`exitCode`) and return its value — optionally with a visible (non-headless)
   mode toggle (wasmbrowsertest's `WASM_HEADLESS=off`) and a WSL
   GPU-disable special case. Check whether `OpenBrowser.go` /
   `devbrowser.go` already has something close enough to extend, versus
   needing a new small exported function — don't build a second one if an
   equivalent already exists.

7. **Coordinate with wasmbrowsertest's own PLAN**
   (`/home/cesar/Dev/Project/tinywasm/wasmbrowsertest/docs/PLAN.md`) once
   this API exists — that repo depends on it to remove its direct chromedp
   import.

## Out of scope

- Filing/fixing the actual Debian `chromium` package bug — that's upstream
  Debian's problem, not ours.
- Rewriting wasmbrowsertest itself — that happens in its own repo, tracked
  in its own PLAN.md, once Part B here is ready to be consumed.
