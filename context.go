package devbrowser

import (
	"errors"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func (h *DevBrowser) CreateBrowserContext() error {
	// Prepare launcher with custom flags similar to previous implementation
	l := launcher.New().Headless(false)
	// Correct usage: pass the flag name (no '=') and the value as a separate argument.
	// Do not include '=' in the flag name — use the name and value parameters instead.
	l.Append("disable-blink-features", "WebFontsInterventionV2")
	l.Append("use-fake-ui-for-media-stream")
	l.Append("no-focus-on-load")
	l.Append("auto-open-devtools-for-tabs")
	l.Append("window-position", h.position)

	u, err := l.Launch()
	if err != nil {
		return errors.New("failed to launch browser: " + err.Error())
	}

	h.launcherURL = u
	h.browser = rod.New().ControlURL(u)
	// Connect to the browser
	if err := h.browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	// Create a new page (use MustPage to simplify creation)
	p := h.browser.MustPage("about:blank")
	h.page = p

	// Cancel function to close resources
	h.cancelFunc = func() {
		if h.page != nil {
			_ = h.page.Close()
		}
		if h.browser != nil {
			_ = h.browser.Close()
		}
		if h.launcherURL != "" {
			// launcher process will exit when browser closed
		}
	}

	return nil
}
