package devbrowser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chromedp/chromedp"
)

// TestBrowserSwipe verifies that the browser_swipe_element tool performs drag actions.
func TestBrowserSwipe(t *testing.T) {
	// 1. Setup a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<style>
				#slider {
					width: 300px;
					height: 20px;
					background: #ccc;
					position: relative;
				}
				#handle {
					width: 20px;
					height: 20px;
					background: red;
					position: absolute;
					left: 0;
					top: 0;
					cursor: pointer;
				}
			</style>
			<body>
				<div id="slider">
					<div id="handle"></div>
				</div>
				<script>
					const handle = document.getElementById('handle');
					let isDragging = false;
					
					handle.addEventListener('mousedown', (e) => {
						isDragging = true;
					});
					
					document.addEventListener('mousemove', (e) => {
						if (isDragging) {
							// Simple horizontal drag logic
							let newLeft = e.clientX - 10; // offset
							if (newLeft < 0) newLeft = 0;
							if (newLeft > 280) newLeft = 280;
							handle.style.left = newLeft + 'px';
						}
					});
					
					document.addEventListener('mouseup', () => {
						isDragging = false;
					});
				</script>
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
	db := &DevBrowser{
		ctx:    ctx,
		cancel: cancel,
		isOpen: true,
		log:    func(args ...any) {},
	}
	db.width = 1024
	db.height = 768

	// Navigate
	if err := chromedp.Run(ctx, chromedp.Navigate(ts.URL)); err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// 4. Get the swipe tool
	tools := db.getInteractionTools()
	var swipeTool ToolMetadata
	for _, tool := range tools {
		if tool.Name == "browser_swipe_element" {
			swipeTool = tool
			break
		}
	}

	if swipeTool.Name == "" {
		t.Fatal("browser_swipe_element tool not found")
	}

	// 5. Execute swipe: Swipe right by 100px on the handle
	swipeTool.Execute(map[string]any{
		"selector":  "#handle",
		"direction": "right",
		"distance":  100.0,
	})

	// 6. Verify handle moved
	// Initial left was 0. Should be ~90-100 depending on exact center math and mouse events.
	// Our mock logic sets left = clientX - 10.
	// Center of handle (20x20) at start (left=0) is x=10.
	// Swipe right 100px -> new mouse x = 110.
	// JS logic: newLeft = 110 - 10 = 100.
	var leftValue string
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.getElementById('handle').style.left`, &leftValue),
	)
	if err != nil {
		t.Fatalf("Failed to check handle position: %v", err)
	}

	if leftValue == "0px" || leftValue == "" {
		t.Errorf("Expected handle to move, but left is %s", leftValue)
	}
}
