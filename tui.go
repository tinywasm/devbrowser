package devbrowser

func (h *DevBrowser) Name() string {
	return "BROWSER"
}

func (h *DevBrowser) Label() string {

	state := "Open Browser"

	if h.isOpen {
		state = "Close Browser"
	}

	return state
}

// run the open/close browser operation
func (h *DevBrowser) Execute(progress chan<- string) {

	if h.isOpen { // cerrar si esta abierto
		progress <- "Closing..."

		if err := h.CloseBrowser(); err != nil {
			progress <- "Close error:" + err.Error()
		} else {
			progress <- "Closed."
		}

	} else { // abrir si esta cerrado
		progress <- "Opening..."
		h.OpenBrowser()

	}

}

// MessageTracker implementation for operation tracking
func (h *DevBrowser) GetLastOperationID() string   { return h.lastOpID }
func (h *DevBrowser) SetLastOperationID(id string) { h.lastOpID = id }
