package devbrowser

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func (h *DevBrowser) OpenBrowser(port string, https bool) {
	h.mu.Lock()
	isFirst := h.firstCall
	h.firstCall = false
	h.lastPort = port
	h.lastHttps = https

	//h.Logger(fmt.Sprintf("DEBUG: OpenBrowser called. port=%s, https=%v, isOpen=%v, firstCall=%v", port, https, h.isOpen, isFirst))

	// Logic: on first call, only open if autoStart is true.
	// On subsequent calls (e.g. user action), always open.
	if isFirst && !h.autoStart {
		//h.Logger("DEBUG: OpenBrowser skipped on first call (autoStart=false)")
		h.mu.Unlock()
		return
	}

	if h.isOpen {
		//h.Logger("DEBUG: OpenBrowser returning early, already open")
		h.mu.Unlock()
		return
	}

	if h.testMode {
		h.openedOnce = true
		h.mu.Unlock()
		h.Logger("Skipping browser open in TestMode")
		return
	}
	h.isOpen = true
	h.openedOnce = true
	h.mu.Unlock()

	// Add listener for exit signal (only once per open session)
	go func() {
		<-h.exitChan
		h.CloseBrowser()
	}()

	go func() {
		err := h.CreateBrowserContext()
		if err != nil {
			h.errChan <- err
			return
		}

		protocol := "http"
		if https {
			protocol = "https"
		}
		url := protocol + `://localhost:` + port + "/"

		// Initialize console log capturing BEFORE navigating to the page
		// This ensures all console.log statements from page load are captured
		if err := h.initializeConsoleCapture(); err != nil {
			h.Logger("Warning: failed to initialize console capture:", err)
			// Continue anyway - capture is optional
		}
		h.initializeNetworkCapture()
		h.initializeErrorCapture()

		if err := chromedp.Run(h.ctx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body"),
		); err != nil {
			h.errChan <- fmt.Errorf("error navigating to %s: %v", url, err)
			return
		}

		// Esperar un momento adicional para asegurar que todo esté cargado
		time.Sleep(100 * time.Millisecond)

		h.readyChan <- true

		// Monitor browser context for manual close
		go h.monitorBrowserClose()
	}()

	// Esperar señal de inicio o error
	select {
	case err := <-h.errChan:
		// use helper to ensure logging goes through configured logger
		h.Logger("Error opening DevBrowser: ", err)
		h.CloseBrowser()
		return
	case <-h.readyChan:
		h.Logger(h.StatusMessage())

		// Start monitoring browser geometry changes
		go h.monitorBrowserGeometry()

		h.ui.RefreshUI()
		return
	}
}
