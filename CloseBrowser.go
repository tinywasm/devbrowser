package devbrowser

import (
	"errors"
)

func (h *DevBrowser) CloseBrowser() error {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if !h.IsOpenFlag {
		return errors.New("DevBrowser is already closed")
	}

	h.IsOpenFlag = false

	if h.Cancel != nil {
		h.Cancel()
	}

	// Limpiar recursos
	h.Ctx = nil
	h.Cancel = nil

	h.Logger(h.StatusMessage())
	h.UI.RefreshUI()
	return nil
}
