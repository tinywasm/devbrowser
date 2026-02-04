package devbrowser

import "github.com/tinywasm/mcpserve"

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestBrowserNavigation verifies that the browser_navigate tool works
func TestBrowserNavigation(t *testing.T) {
	// 1. Setup a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Navigated!</h1></body></html>`)
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

	// 3. Create a devbrowser instance with this context
	db := &DevBrowser{
		ctx:    ctx,
		cancel: cancel,
		isOpen: true,
		log:    func(args ...any) {}, // No-op logger
	}
	db.width = 1024
	db.height = 768

	// 4. Get the navigation tool
	tools := db.getNavigationTools()
	var navigateTool mcpserve.ToolMetadata
	for _, tool := range tools {
		if tool.Name == "browser_navigate" {
			navigateTool = tool
			break
		}
	}

	if navigateTool.Name == "" {
		t.Fatal("browser_navigate tool not found")
	}

	// 5. Execute the tool
	navigateTool.Execute(map[string]any{
		"url": ts.URL,
	})

	// 6. Verify navigation happened by checking the page content
	var content string
	err := chromedp.Run(ctx,
		chromedp.Sleep(200*time.Millisecond),
		chromedp.OuterHTML("h1", &content, chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("Failed to get page content: %v", err)
	}

	if !strings.Contains(content, "Navigated!") {
		t.Errorf("Expected content to contain 'Navigated!', got '%s'", content)
	}
}
