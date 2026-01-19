package devbrowser

import (
	"errors"
)

func (h *DevBrowser) CloseBrowser() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isOpen {
		return errors.New("DevBrowser is already closed")
	}

	h.isOpen = false

	if h.cancel != nil {
		h.cancel()
	}

	// Limpiar recursos
	h.ctx = nil
	h.cancel = nil

	h.Logger(h.StatusMessage())
	h.ui.RefreshUI()
	return nil
}
