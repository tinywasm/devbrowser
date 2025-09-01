package devbrowser

import (
	"errors"
)

func (h *DevBrowser) CloseBrowser() error {
	if !h.isOpen {
		return errors.New("DevBrowser is already closed")
	}

	// Llamar a la función de cancelación personalizada que cierra todos los recursos
	if h.cancelFunc != nil {
		h.cancelFunc()
		h.isOpen = false
	}

	// Limpiar recursos
	h.launcherURL = ""
	h.browser = nil
	h.page = nil
	h.cancelFunc = nil

	return nil
}
