package devbrowser

import "github.com/tinywasm/mcpserve"

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tinywasm/devbrowser/chromedp"
)

// TestRobustClickVerification verifies that browser_click_element falls back to JS click
// when standard click fails (e.g., element is invisible).
func TestRobustClickVerification(t *testing.T) {
	// 1. Setup a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<body>
				<!-- Hidden button (display: none) to force Click() timeout/failure -->
				<button id="ghost-btn" style="display: none;" onclick="document.body.innerHTML += '<div id=result>Clicked!</div>'">Ghost Click</button>
			</body>
			</html>
		`)
	}))
	defer ts.Close()

	// 2. Setup Chromedp
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.DisableGPU,
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 3. Create a devbrowser instance
	logs := []string{}
	db := &DevBrowser{
		ctx:    ctx,
		cancel: cancel,
		isOpen: true,
		log: func(args ...any) {
			logs = append(logs, fmt.Sprint(args...))
		},
	}
	db.width = 1024
	db.height = 768

	// Navigate to page
	if err := chromedp.Run(ctx, chromedp.Navigate(ts.URL)); err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// 4. Get the interaction tool
	tools := db.getInteractionTools()
	var clickTool mcpserve.ToolMetadata
	for _, tool := range tools {
		if tool.Name == "browser_click_element" {
			clickTool = tool
			break
		}
	}

	// 5. Execute the tool on the invisible button
	// Standard click should timeout (it waits for visibility), then JS click should trigger.
	clickTool.Execute(map[string]any{
		"selector": "#ghost-btn",
		"timeout":  3000.0, // Enough for 1s fallback delay + execution
	})

	// 6. Verify result
	var resultText string
	err := chromedp.Run(ctx,
		chromedp.Text("#result", &resultText, chromedp.ByQuery),
	)

	if err != nil {
		t.Fatalf("Failed to verify click result (JS fallback likely failed): %v", err)
	}

	if resultText != "Clicked!" {
		t.Errorf("Expected result 'Clicked!', got '%s'", resultText)
	}

	// Logs confirm fallback occurred (verified manually via test output if needed)
}
