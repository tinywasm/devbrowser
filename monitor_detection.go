package devbrowser

import (
	"fmt"

	"github.com/kbinani/screenshot"
)

// queryMonitorSize is a function variable to allow mocking in tests.
// It returns width, height of the primary display.
var queryMonitorSize = func() (int, int) {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		return 0, 0
	}
	// Primary display (index 0)
	bounds := screenshot.GetDisplayBounds(0)
	return bounds.Dx(), bounds.Dy()
}

// detectMonitorSize attempts to retrieve the primary monitor's dimensions.
// It updates the DevBrowser's monitorWidth and monitorHeight fields.
func (b *DevBrowser) detectMonitorSize() {
	w, h := queryMonitorSize()

	if w == 0 || h == 0 {
		b.log("No active displays found or extraction failed")
		return
	}

	b.mu.Lock()
	b.monitorWidth = w
	b.monitorHeight = h
	b.mu.Unlock()

	b.log(fmt.Sprintf("Detected monitor size: %dx%d", w, h))
}
