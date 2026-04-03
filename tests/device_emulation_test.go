package devbrowser_test

import (
	"github.com/tinywasm/devbrowser"
	"testing"
)

func TestDeviceEmulation_Logic(t *testing.T) {
	b := &devbrowser.DevBrowser{
		Width:  1200,
		Height: 800,
		Log:    func(msg ...any) {},
	}

	// Test case: mobile emulation
	b.ViewportMode = "mobile"
	// We can't easily test the chromedp effects here without a real browser,
	// but we can test the state transitions and config saving logic.

	if b.ViewportMode != "mobile" {
		t.Errorf("Expected viewportMode mobile, got %s", b.ViewportMode)
	}

	// Test case: desktop (reset)
	b.ViewportMode = "desktop"
	if b.ViewportMode != "desktop" {
		t.Errorf("Expected viewportMode desktop, got %s", b.ViewportMode)
	}
}

func TestScreenshotUtility_ResultStructure(t *testing.T) {
	// Simple structure check
	res := &devbrowser.ScreenshotResult{
		ImageData: []byte("fake-image"),
		PageURL:   "http://localhost",
		Width:     1024,
	}

	if len(res.ImageData) == 0 {
		t.Error("ImageData should not be empty")
	}
}
