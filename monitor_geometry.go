package devbrowser

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/tinywasm/devbrowser/cdproto/browser"
	"github.com/tinywasm/devbrowser/chromedp"
)

// monitorBrowserGeometry monitors changes in browser window position and size
// and automatically saves them to the database
func (b *DevBrowser) monitorBrowserGeometry() {
	b.Mu.Lock()
	ctx := b.Ctx
	isOpen := b.IsOpenFlag
	b.Mu.Unlock()

	if ctx == nil || !isOpen {
		return
	}

	// Wait a bit before starting to monitor to allow Chrome to stabilize
	// after initial window creation and avoid capturing transient sizes
	time.Sleep(3 * time.Second)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Browser context closed, stop monitoring
			return
		case <-ticker.C:
			// Check and save current geometry if changed
			b.checkAndSaveGeometry()
		}
	}
}

// checkAndSaveGeometry checks current browser geometry and saves if changed
func (b *DevBrowser) checkAndSaveGeometry() {
	b.Mu.Lock()
	ctx := b.Ctx
	b.Mu.Unlock()

	if ctx == nil {
		return
	}

	var x, y, width, height int64

	// Get window bounds using Chrome DevTools Protocol
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Get the window ID for the current target
			t := chromedp.FromContext(ctx).Target

			// Get window bounds for target
			windowID, bounds, err := browser.GetWindowForTarget().WithTargetID(t.TargetID).Do(ctx)
			if err != nil {
				return err
			}

			// windowID is returned but we use bounds directly
			_ = windowID

			x = bounds.Left
			y = bounds.Top
			width = bounds.Width
			height = bounds.Height
			return nil
		}),
	)

	if err != nil {
		// Silently ignore errors - geometry monitoring is non-critical
		return
	}

	// Check if position changed
	newPosition := strconv.FormatInt(x, 10) + "," + strconv.FormatInt(y, 10)
	if newPosition != b.Position {
		b.Position = newPosition
		b.DB.Set(StoreKeyBrowserPosition, b.Position)
	}

	// Check if width or height changed
	newWidth := int(width)
	newHeight := int(height)

	if (newWidth != b.Width && newWidth > 0) || (newHeight != b.Height && newHeight > 0) {
		b.Width = newWidth
		b.Height = newHeight
		b.SizeConfigured = true // Mark as manually configured

		size := fmt.Sprintf("%d,%d", b.Width, b.Height)
		b.DB.Set(StoreKeyBrowserSize, size)
	}
}
