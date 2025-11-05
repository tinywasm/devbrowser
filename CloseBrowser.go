package devbrowser

import (
	"errors"
)

func (h *DevBrowser) CloseBrowser() error {
	if !h.isOpen {
		return errors.New("DevBrowser is already closed")
	}

	if h.cancel != nil {
		h.cancel()
		h.isOpen = false
	}

	// Limpiar recursos
	h.ctx = nil
	h.cancel = nil

	return nil
}
