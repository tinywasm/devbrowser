package devbrowser

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestScreenshotViewport(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getScreenshotTools()[0].Execute(map[string]any{"fullpage": false}, progress)

	output := <-progress
	if !strings.HasPrefix(output, "data:image/png;base64,") {
		t.Errorf("Expected base64 PNG data URI, got %s", output)
	}

	b64data := strings.TrimPrefix(output, "data:image/png;base64,")
	decoded, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		t.Errorf("Failed to decode base64: %v", err)
	}

	if len(decoded) == 0 {
		t.Error("Decoded PNG is empty")
	}
}

func TestScreenshotFullPage(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	// Create a long page to test fullpage screenshot
	script := `document.body.style.height = '2000px';`
	var res interface{}
	err := chromedp.Run(db.ctx, chromedp.Evaluate(script, &res))
	if err != nil {
		t.Fatalf("Failed to set page height: %v", err)
	}

	progress := make(chan string)
	go db.getScreenshotTools()[0].Execute(map[string]any{"fullpage": true}, progress)

	output := <-progress
	if !strings.HasPrefix(output, "data:image/png;base64,") {
		t.Errorf("Expected base64 PNG data URI, got %s", output)
	}
}

func TestScreenshotBrowserNotOpen(t *testing.T) {
	db, _ := DefaultTestBrowser()
	// Do not open browser

	progress := make(chan string)
	go db.getScreenshotTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	if output != "Browser is not open. Please open it first with browser_open" {
		t.Errorf("Expected 'Browser is not open' error, got '%s'", output)
	}
}
