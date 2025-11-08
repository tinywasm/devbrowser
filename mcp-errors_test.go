package devbrowser

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestCaptureJavaScriptError(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	// Execute script that throws an error
	script := `throw new Error('test error')`
	var res interface{}
	// We don't care about the result, just that it executes
	go chromedp.Run(db.ctx, chromedp.Evaluate(script, &res))

	// Allow some time for the error to be captured
	time.Sleep(100 * time.Millisecond)

	progress := make(chan string)
	go db.getErrorTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	if !strings.Contains(output, "test error") {
		t.Errorf("Expected error message containing 'test error', got '%s'", output)
	}
}

func TestNoErrorsWhenNone(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getErrorTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	if output != "No JavaScript errors captured" {
		t.Errorf("Expected 'No JavaScript errors captured', got '%s'", output)
	}
}

func TestErrorsBrowserNotOpen(t *testing.T) {
	db, _ := DefaultTestBrowser()

	progress := make(chan string)
	go db.getErrorTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	if output != "Browser is not open. Please open it first with browser_open" {
		t.Errorf("Expected 'Browser is not open' error, got '%s'", output)
	}
}
