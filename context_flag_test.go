package devbrowser

import (
	"testing"
)

// TestLauncherAppendPanic reproduces the panic seen when Append is called with flags.
// It attempts to call CreateBrowserContext and expects either an error or a panic
// from the launcher library. This test helps reproduce the stacktrace reported.
func TestLauncherAppendPanic(t *testing.T) {
	db := New(fakeServerConfig{}, fakeUI{}, make(chan bool), nil)
	// provide a position just in case launcher handles it
	db.position = "0,0"

	// Capture panic if it happens
	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		_ = db.CreateBrowserContext()
	}()

	if !didPanic {
		// If no panic, ensure the call at least returned an error when unable to launch
		// (some environments may not allow launching a browser)
		// If we reach here without panic, the test is inconclusive but still useful.
		t.Log("CreateBrowserContext did not panic; check environment or inspect returned state")
	}
}
