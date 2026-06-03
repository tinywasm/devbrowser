package devbrowser

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/tinywasm/devbrowser/cdproto/browser"
	"github.com/tinywasm/devbrowser/chromedp"
)

func (h *DevBrowser) BrowserPositionAndSizeChanged(fieldName string, oldValue, newValue string) error {

	if !h.IsOpenFlag {
		return nil
	}

	err := h.SetBrowserPositionAndSize(newValue)
	if err != nil {
		return err
	}

	return h.RestartBrowser()
}

func (b *DevBrowser) SetBrowserPositionAndSize(newConfig string) (err error) {

	this := errors.New("setBrowserPositionAndSize")

	position, width, height, err := getBrowserPositionAndSize(newConfig)

	if err != nil {
		return errors.Join(this, err)
	}
	b.Position = position

	b.Width, err = strconv.Atoi(width)
	if err != nil {
		return errors.Join(this, err)
	}
	b.Height, err = strconv.Atoi(height)
	if err != nil {
		return errors.Join(this, err)
	}

	// Save the new configuration to db
	if err := b.SaveConfig(); err != nil {
		return errors.Join(this, err)
	}

	return
}

// applyConfiguredPosition forces the browser window to the configured position
// via CDP SetWindowBounds. Called after ReadyChan so the context is ready.
// This overrides the WM placement (--window-position is a hint that WMs may ignore).
func (b *DevBrowser) applyConfiguredPosition() {
	b.Mu.Lock()
	ctx := b.Ctx
	pos := b.Position
	b.Mu.Unlock()

	if ctx == nil || pos == "" || pos == "0,0" {
		return
	}

	parts := strings.SplitN(pos, ",", 2)
	if len(parts) != 2 {
		return
	}
	x, err1 := strconv.ParseInt(parts[0], 10, 64)
	y, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		return
	}

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			t := chromedp.FromContext(ctx).Target
			windowID, _, err := browser.GetWindowForTarget().WithTargetID(t.TargetID).Do(ctx)
			if err != nil {
				return err
			}
			return browser.SetWindowBounds(windowID, &browser.Bounds{
				Left:        x,
				Top:         y,
				WindowState: browser.WindowStateNormal,
			}).Do(ctx)
		}),
	)
	if err != nil {
		b.Logger("Warning: could not apply configured position:", err)
	}
}

func getBrowserPositionAndSize(config string) (position, width, height string, err error) {
	current := strings.Split(config, ":")

	if len(current) != 2 {
		err = errors.New("browse config must be in the format: 1930,0:800,600")
		return
	}

	positions := strings.Split(current[0], ",")
	if len(positions) != 2 {
		err = errors.New("position must be with commas e.g.: 1930,0:800,600")
		return
	}
	position = current[0]

	sizes := strings.Split(current[1], ",")
	if len(sizes) != 2 {
		err = errors.New("width and height must be with commas e.g.: 1930,0:800,600")
		return
	}

	widthInt, err := strconv.Atoi(sizes[0])
	if err != nil {
		err = errors.New("width must be an integer number")
		return
	}
	width = strconv.Itoa(widthInt)

	heightInt, err := strconv.Atoi(sizes[1])
	if err != nil {
		err = errors.New("height must be an integer number")
		return
	}
	height = strconv.Itoa(heightInt)

	return
}
