package devbrowser

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

func (h *DevBrowser) OpenBrowser() {
	if h.isOpen {
		return
	}

	// Add listener for exit signal
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

		h.isOpen = true
		var protocol = "http"
		url := protocol + `://localhost:` + h.config.ServerPort() + "/"

		// Initialize console log capturing BEFORE navigating to the page
		// This ensures all console.log statements from page load are captured
		if err := h.initializeConsoleCapture(); err != nil {
			h.logger("Warning: failed to initialize console capture:", err)
			// Continue anyway - capture is optional
		}

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
		h.isOpen = false
		// use helper to ensure logging goes through configured logger
		h.logger("Error opening DevBrowser: ", err)
		return
	case <-h.readyChan:
		// Tomar el foco de la UI después de abrir el navegador
		/*  err := h.ui.ReturnFocus()
		if err != nil {
			h.logger.Write([]byte("Error returning focus to UI: " + err.Error()))
		} */
		h.logger("Started")

		h.ui.RefreshUI()
		return
	}
}
