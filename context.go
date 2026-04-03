package devbrowser

import (
	"context"
	"strings"

	"github.com/tinywasm/devbrowser/chromedp"
)

func (h *DevBrowser) CreateBrowserContext() error {
	// Create allocator with custom options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", h.Headless),
		chromedp.Flag("disable-blink-features", "WebFontsInterventionV2"),
		chromedp.Flag("use-fake-ui-for-media-stream", true),
		chromedp.Flag("window-position", h.Position),
		chromedp.WindowSize(h.Width, h.Height),
	)

	// Conditionally add devtools flag
	if h.Width > 1200 {
		opts = append(opts, chromedp.Flag("auto-open-devtools-for-tabs", true))
	}

	// Disable cache by default unless explicitly enabled
	// Note: disk-cache-size and media-cache-size flags cause "invalid exec pool flag" errors
	// Use --disable-cache instead
	if !h.CacheEnabled {
		opts = append(opts,
			chromedp.Flag("disable-cache", true),
			chromedp.Flag("disable-gpu-shader-disk-cache", true),
		)
	}

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx,
		chromedp.WithErrorf(func(format string, args ...any) {
			// Chrome sends new CDP enum values before cdproto is updated.
			// These unmarshal errors are harmless — suppress them to avoid
			// corrupting the TUI via stderr.
			if strings.HasPrefix(format, "could not unmarshal event") {
				return
			}
			errorArgs := append([]any{"ERROR: "}, args...)
			// Forward error to devbrowser log
			h.Log(errorArgs...)
		}),
	)
	h.Ctx = ctx
	h.Cancel = cancel

	return nil
}
