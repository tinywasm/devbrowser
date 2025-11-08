package devbrowser

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

func TestClickButton(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	// Add a button to the page
	script := `
		const btn = document.createElement('button');
		btn.id = 'test-btn';
		btn.innerText = 'Click Me';
		btn.onclick = () => { document.body.style.backgroundColor = 'red'; };
		document.body.appendChild(btn);
	`
	var res interface{}
	err := chromedp.Run(db.ctx, chromedp.Evaluate(script, &res))
	if err != nil {
		t.Fatalf("Failed to add button: %v", err)
	}

	progress := make(chan string)
	go db.getInteractionTools()[0].Execute(map[string]any{"selector": "#test-btn"}, progress)

	output := <-progress
	if output != "Clicked element: #test-btn" {
		t.Errorf("Expected 'Clicked element: #test-btn', got '%s'", output)
	}

	// Verify the button was clicked
	var bgColor string
	err = chromedp.Run(db.ctx, chromedp.Evaluate(`document.body.style.backgroundColor`, &bgColor))
	if err != nil {
		t.Fatalf("Failed to get background color: %v", err)
	}
	if bgColor != "red" {
		t.Errorf("Expected background color to be red, got %s", bgColor)
	}
}

func TestClickElementNotFound(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getInteractionTools()[0].Execute(map[string]any{"selector": "#nonexistent"}, progress)

	output := <-progress
	if !strings.HasPrefix(output, "Error clicking element") {
		t.Errorf("Expected error message, got '%s'", output)
	}
}

func TestClickBrowserNotOpen(t *testing.T) {
	db, _ := DefaultTestBrowser()

	progress := make(chan string)
	go db.getInteractionTools()[0].Execute(map[string]any{"selector": "#test"}, progress)

	output := <-progress
	if output != "Browser is not open. Please open it first with browser_open" {
		t.Errorf("Expected 'Browser is not open' error, got '%s'", output)
	}
}
