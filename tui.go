package devbrowser

import "sync/atomic"

func (h *DevBrowser) Name() string {
	return "BROWSER"
}

func (h *DevBrowser) Label() string {
	return "Auto Start Browser 't/f'"
}

// Value returns current auto-start setting as "t" or "f"
func (h *DevBrowser) Value() string {
	if val, err := h.db.Get(StoreKeyBrowserAutostart); err == nil && val != "" {
		if val == "true" {
			return "t"
		}
		return "f"
	}
	return "t" // Default: auto-start enabled
}

// StatusMessage returns formatted browser status for logging
// Format: "Open | Auto-Start: t | Shortcut B" or "Closed | Auto-Start: f | Shortcut B"
func (h *DevBrowser) StatusMessage() string {
	state := "Closed"
	if h.isOpen {
		state = "Open"
	}
	return state + " | Auto-Start: " + h.Value() + " | Shortcut B"
}

// Change handles user input: toggles auto-start or browser state
func (h *DevBrowser) Change(newValue string) {
	switch newValue {
	case "B": // Shortcut: toggle browser open/close
		go func() {
			if !atomic.CompareAndSwapInt32(&h.busy, 0, 1) {
				// Prevent spamming / re-entrant calls
				return
			}
			defer atomic.StoreInt32(&h.busy, 0)

			if h.isOpen {
				if err := h.CloseBrowser(); err != nil {
					h.Logger("Close error:", err.Error())
				}
			} else {
				h.OpenBrowser()
			}
			// Note: OpenBrowser and CloseBrowser log StatusMessage internally
		}()
		return // Return immediately, don't fall through to RefreshUI below

	default: // Toggle auto-start setting
		current := h.Value()
		newVal := "true"
		if current == "t" {
			newVal = "false"
		}
		h.db.Set(StoreKeyBrowserAutostart, newVal)
		h.Logger(h.StatusMessage())
	}

	h.ui.RefreshUI()
}

// Shortcuts registers "B" for browser toggle
func (h *DevBrowser) Shortcuts() []map[string]string {
	return []map[string]string{
		{"B": "toggle browser"},
	}
}
