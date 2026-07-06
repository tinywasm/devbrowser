package devbrowser

import (
	"context"
	"fmt"

	"github.com/tinywasm/devbrowser/chromedp"
)

// RunWasmTest launches a browser, navigates to the given URL, waits for the
// #doneButton to be enabled, and returns the value of the exitCode global.
// This is designed to be used by wasmbrowsertest.
func RunWasmTest(ctx context.Context, url string, headless bool) (int, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.ExecPath(ResolveChromeExecPath()),
	)

	// In some environments (like WSL), GPU can cause issues in headless mode.
	// We include DisableGPU by default as it's a common requirement for WASM tests.
	opts = append(opts, chromedp.DisableGPU)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var exitCode int
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(url),
		// Wait for #doneButton to exist and NOT have the disabled attribute
		chromedp.WaitVisible(`#doneButton:not([disabled])`, chromedp.ByQuery),
		// Evaluate the exitCode global variable
		chromedp.Evaluate(`window.exitCode`, &exitCode),
	)

	if err != nil {
		return 0, fmt.Errorf("wasm test failed: %w", err)
	}

	return exitCode, nil
}
