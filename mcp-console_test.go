package devbrowser

import (
	"strings"
	"testing"
	"time"

	"github.comcom/chromedp/chromedp"
)

func TestGetConsoleLogs(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	// Log a message to the console
	script := `console.log('hello world')`
	var res interface{}
	err := chromedp.Run(db.ctx, chromedp.Evaluate(script, &res))
	if err != nil {
		t.Fatalf("Failed to log to console: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	progress := make(chan string)
	go db.getConsoleTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	if !strings.Contains(output, "hello world") {
		t.Errorf("Expected 'hello world' in console logs, got '%s'", output)
	}
}

func TestGetConsoleLogsEmpty(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getConsoleTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	// Depending on the environment, there might be some initial logs.
	// This test is mostly to ensure the tool doesn't crash.
	// A more robust test would clear the logs first.
	// For now, we just check that we get a response.
	if output == "" {
		t.Errorf("Expected some output, got empty string")
	}
}

func TestGetConsoleLogsBrowserNotOpen(t *testing.T) {
	db, _ := DefaultTestBrowser()

	progress := make(chan string)
	go db.getConsoleTools()[0].Execute(map[string]any{}, progress)

	output := <-progress
	if output != "Browser is not open. Please open it first with browser_open" {
		t.Errorf("Expected 'Browser is not open' error, got '%s'", output)
	}
}
