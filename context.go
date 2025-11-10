package devbrowser

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

func (h *DevBrowser) CreateBrowserContext() error {
	// Format window size as "width,height"
	windowSize := fmt.Sprintf("%d,%d", h.width, h.height)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", h.headless),
		chromedp.Flag("disable-blink-features", "WebFontsInterventionV2"),
		chromedp.Flag("use-fake-ui-for-media-stream", true),
		chromedp.Flag("no-focus-on-load", true),
		chromedp.Flag("auto-open-devtools-for-tabs", true),
		chromedp.Flag("window-position", h.position),
		chromedp.Flag("window-size", windowSize),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	// Adapter function to match the logger signature required by chromedp.WithLogf.
	logfAdapter := func(format string, args ...any) {
		h.logger(fmt.Sprintf(format, args...))
	}

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logfAdapter))
	h.ctx = ctx
	h.cancel = cancel

	return nil
}
