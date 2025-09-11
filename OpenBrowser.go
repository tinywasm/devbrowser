package devbrowser

import (
	"fmt"
	"time"
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
		url := protocol + `://localhost:` + h.config.GetServerPort() + "/"

		// Navegar a la URL (rod)
		if h.page == nil {
			h.errChan <- fmt.Errorf("page not initialized")
			return
		}

		if err = h.page.Navigate(url); err != nil {
			h.errChan <- fmt.Errorf("error navigating to %s: %v", url, err)
			return
		}

		// Esperar carga completa usando rod
		if err = h.page.WaitLoad(); err != nil {
			h.errChan <- fmt.Errorf("error waiting for load: %v", err)
			return
		}

		// Esperar un momento adicional para asegurar que todo esté cargado
		time.Sleep(100 * time.Millisecond)
		h.readyChan <- true
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
		return
	}
}
