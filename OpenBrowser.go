package devbrowser

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func (h *DevBrowser) OpenBrowser() {
	h.mu.Lock()
	if h.isOpen {
		h.mu.Unlock()
		return
	}

	if h.testMode || TestMode {
		h.mu.Unlock()
		h.Logger("Skipping browser open in TestMode")
		return
	}
	h.isOpen = true
	h.mu.Unlock()

	// Add listener for exit signal (only once per open session)
	go func() {
		<-h.exitChan
		h.CloseBrowser()
	}()
	// fmt.Println("*** START DEV BROWSER ***")
	go func() {
		err := h.CreateBrowserContext()
		if err != nil {
			h.errChan <- err
			return
		}

		var protocol = "http"
		url := protocol + `://localhost:` + h.config.ServerPort() + "/"

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
		// Tomar el foco de la UI después de abrir el navegador
		/*  err := h.ui.ReturnFocus()
		if err != nil {
			h.Logger.Write([]byte("Error returning focus to UI: " + err.Error()))
		} */
		h.Logger(h.StatusMessage())

		// Start monitoring browser geometry changes
		go h.monitorBrowserGeometry()

		h.ui.RefreshUI()
		return
	}
}
