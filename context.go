package devbrowser

import (
	"context"

	"github.com/chromedp/chromedp"
)

func (h *DevBrowser) CreateBrowserContext() error {
	// Create allocator with custom options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", h.headless),
		chromedp.Flag("disable-blink-features", "WebFontsInterventionV2"),
		chromedp.Flag("use-fake-ui-for-media-stream", true),
		chromedp.Flag("window-position", h.position),
		chromedp.WindowSize(h.width, h.height),
	)

	// Conditionally add devtools flag
	if h.width > 1200 {
		opts = append(opts, chromedp.Flag("auto-open-devtools-for-tabs", true))
	}

	// Disable cache by default unless explicitly enabled
	// Note: disk-cache-size and media-cache-size flags cause "invalid exec pool flag" errors
	// Use --disable-cache instead
	if !h.cacheEnabled {
		opts = append(opts,
			chromedp.Flag("disable-cache", true),
			chromedp.Flag("disable-gpu-shader-disk-cache", true),
		)
	}

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	h.ctx = ctx
	h.cancel = cancel

	return nil
}
