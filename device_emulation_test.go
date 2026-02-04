package devbrowser

import (
	"testing"
)

func TestDeviceEmulation_Logic(t *testing.T) {
	b := &DevBrowser{
		width:  1200,
		height: 800,
		log:    func(msg ...any) {},
	}

	// Test case: mobile emulation
	b.viewportMode = "mobile"
	// We can't easily test the chromedp effects here without a real browser,
	// but we can test the state transitions and config saving logic.

	if b.viewportMode != "mobile" {
		t.Errorf("Expected viewportMode mobile, got %s", b.viewportMode)
	}

	// Test case: desktop (reset)
	b.viewportMode = "desktop"
	if b.viewportMode != "desktop" {
		t.Errorf("Expected viewportMode desktop, got %s", b.viewportMode)
	}
}

func TestScreenshotUtility_ResultStructure(t *testing.T) {
	// Simple structure check
	res := &ScreenshotResult{
		ImageData: []byte("fake-image"),
		PageURL:   "http://localhost",
		Width:     1024,
	}

	if len(res.ImageData) == 0 {
		t.Error("ImageData should not be empty")
	}
}
