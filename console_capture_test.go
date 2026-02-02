package devbrowser

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestConsoleCapture(t *testing.T) {
	// 1. Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/404.png" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
			<!DOCTYPE html>
			<html>
			<body>
				<script>
					console.log("Normal log message");
					// Throw error after a short delay
					setTimeout(() => {
						throw new Error("Boom - Uncaught Exception");
					}, 50);
				</script>
				<!-- Trigger network error -->
				<img src="/404.png"> 
			</body>
			</html>
		`)
	}))
	defer ts.Close()

	// 2. Initialize DevBrowser Context manually
	db, _ := DefaultTestBrowser()

	if err := db.CreateBrowserContext(); err != nil {
		t.Fatalf("Failed to create browser context: %v", err)
	}
	// We need to initialize capture manually since we aren't using OpenBrowser
	// initializeConsoleCapture is private (lowercase), but we are in the same package `devbrowser`
	// so we can call it.
	if err := db.initializeConsoleCapture(); err != nil {
		t.Fatalf("Failed to init capture: %v", err)
	}

	defer db.CloseBrowser()

	// 3. Navigate to test page and wait for events
	err := chromedp.Run(db.ctx,
		chromedp.Navigate(ts.URL),
		chromedp.Sleep(1000*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// 4. Get Logs
	logs, err := db.GetConsoleLogs()
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	t.Logf("Captured Logs:\n%v", strings.Join(logs, "\n"))

	// 5. Verify Content
	foundNormal := false
	foundException := false
	foundNetwork := false

	for _, log := range logs {
		if strings.Contains(log, "Normal log message") {
			foundNormal = true
		}
		if strings.Contains(log, "Uncaught Exception") && strings.Contains(log, "Boom") {
			foundException = true
		}
		// Network errors come as [error] Network: ... or similar depending on browser version/execution
		// The helper formats it as: [Level] Source: Text
		if strings.Contains(log, "404") || strings.Contains(log, "Failed to load resource") {
			foundNetwork = true
		}
	}

	if !foundNormal {
		t.Error("Did not capture 'Normal log message'")
	}
	if !foundException {
		t.Error("Did not capture 'Uncaught Exception'")
	}
	if !foundNetwork {
		t.Error("Did not capture Network/404 error")
	}
}
