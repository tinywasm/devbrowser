package devbrowser

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestNetworkLogsCapture(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	// Allow some time for initial requests to complete
	time.Sleep(200 * time.Millisecond)

	progress := make(chan string)
	go db.getNetworkTools()[0].Execute(map[string]any{}, progress)

	// We expect multiple log entries, let's just check for one
	output := <-progress
	if !strings.Contains(output, "http") {
		t.Errorf("Expected network log entry, got '%s'", output)
	}
}

func TestNetworkLogsXHRFilter(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	// Make an XHR request
	script := `fetch('/test')`
	var res interface{}
	err := chromedp.Run(db.ctx, chromedp.Evaluate(script, &res))
	if err != nil {
		t.Fatalf("Failed to make fetch request: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	progress := make(chan string)
	go db.getNetworkTools()[0].Execute(map[string]any{"filter": "fetch"}, progress)

	output := <-progress
	if !strings.Contains(output, "[fetch]") {
		t.Errorf("Expected fetch log entry, got '%s'", output)
	}
}

func TestNetworkLogsBrowserNotOpen(t *testing.T) {
	db, _ := DefaultTestBrowser()

	progress := make(chan string)
	go db.getNetworkTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	if output != "Browser is not open. Please open it first with browser_open" {
		t.Errorf("Expected 'Browser is not open' error, got '%s'", output)
	}
}
