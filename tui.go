package devbrowser

import "sync/atomic"

const (
	autoStartOn           = "on"
	autoStartOff          = "off"
	shortcutBrowserToggle = "B"
)

func (h *DevBrowser) Name() string {
	return "BROWSER"
}

func (h *DevBrowser) Label() string {
	return "Auto Start"
}

// Options returns the available auto-start options
func (h *DevBrowser) Options() []map[string]string {
	return []map[string]string{
		{autoStartOn: "On"},
		{autoStartOff: "Off"},
	}
}

// Value returns current auto-start setting as "on" or "off"
func (h *DevBrowser) Value() string {
	if h.AutoStart {
		return autoStartOn
	}
	return autoStartOff
}

// StatusMessage returns formatted browser status for logging
// Format: "Open | Auto-Start: on | Shortcut B" or "Closed | Auto-Start: off | Shortcut B"
func (h *DevBrowser) StatusMessage() string {
	state := "Closed"
	if h.IsOpenFlag {
		state = "Open"
	}
	return state + " | Auto-Start: " + h.Value() + " | Shortcut B"
}

// Change handles user input: sets auto-start or toggles browser state
func (h *DevBrowser) Change(newValue string) {
	switch newValue {
	case shortcutBrowserToggle: // Shortcut: toggle browser open/close
		go func() {
			if !atomic.CompareAndSwapInt32(&h.Busy, 0, 1) {
				// Prevent spamming / re-entrant calls
				return
			}
			defer atomic.StoreInt32(&h.Busy, 0)

			if h.IsOpenFlag {
				if err := h.CloseBrowser(); err != nil {
					h.Logger("Close error:", err.Error())
				}
			} else {
				h.OpenBrowser(h.LastPort, h.LastHttps)
			}
			// Note: OpenBrowser and CloseBrowser log StatusMessage internally
		}()
		return // Return immediately, don't fall through to RefreshUI below

	case autoStartOn:
		h.AutoStart = true
	case autoStartOff:
		h.AutoStart = false

	default:
		h.Logger("Unknown browser setting:", newValue)
		return
	}

	h.SaveConfig()
	h.Logger(h.StatusMessage())
	h.UI.RefreshUI()
}

// Shortcuts registers "B" for browser toggle
func (h *DevBrowser) Shortcuts() []map[string]string {
	return []map[string]string{
		{"B": "toggle browser"},
	}
}
