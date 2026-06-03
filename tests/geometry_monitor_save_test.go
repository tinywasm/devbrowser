package devbrowser_test

import (
	"testing"

	"github.com/tinywasm/devbrowser"
)

// TestPositionNotOverwrittenWhenCDPReportsZero verifies the safety-net guard:
// if CDP transiently reports x=0, y=0 (e.g. right after launch, before
// applyConfiguredPosition has moved the window), a configured non-zero position
// must NOT be overwritten with 0,0.
// Previously FAILED because the position block had no zero-value guard.
func TestPositionNotOverwrittenWhenCDPReportsZero(t *testing.T) {
	store := &defaultStore{m: map[string]string{
		"browser_position": "300,400",
		"browser_size":     "800,600",
	}}
	exitChan := make(chan bool)
	b := devbrowser.New(defaultUI{}, store, exitChan)
	// After New(): b.Position="300,400", b.Width=800, b.Height=600

	// CDP reports x=0, y=0 (transient post-launch) with a different size.
	b.SaveGeometry(0, 0, 1024, 768)

	savedPos, _ := store.Get("browser_position")
	savedSize, _ := store.Get("browser_size")

	if savedSize != "1024,768" {
		t.Errorf("size not saved: got %q, want %q", savedSize, "1024,768")
	}

	if savedPos != "300,400" {
		t.Errorf("BUG: configured position %q overwritten with CDP-reported %q. "+
			"Position block needs x > 0 || y > 0 guard.",
			"300,400", savedPos)
	}
}

// TestPositionSavedWhenSizeChanges verifies that when the window is resized,
// the current position is also written to the store in the same operation,
// even when the position value did not change.
// Previously FAILED because size and position were saved in independent blocks.
func TestPositionSavedWhenSizeChanges(t *testing.T) {
	// Store starts empty — position was never persisted yet.
	store := &defaultStore{m: map[string]string{}}
	exitChan := make(chan bool)
	b := devbrowser.New(defaultUI{}, store, exitChan)

	// Position was applied programmatically (e.g. SetWindowBounds) but not yet in store.
	b.Position = "300,400"

	// CDP correctly reports the applied position and a new size (user resized window).
	b.SaveGeometry(300, 400, 1200, 700)

	savedSize, _ := store.Get("browser_size")
	if savedSize != "1200,700" {
		t.Errorf("size not saved: got %q, want %q", savedSize, "1200,700")
	}

	savedPos, err := store.Get("browser_position")
	if err != nil || savedPos != "300,400" {
		t.Errorf("BUG: position %q not saved when size changed (err: %v). "+
			"When size changes, position must be saved in the same block.",
			savedPos, err)
	}
}
