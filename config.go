package devbrowser

import "strconv"

// Store keys for browser configuration
const (
	StoreKeyBrowserAutostart = "browser_autostart"
	StoreKeyBrowserPosition  = "browser_position"
	StoreKeyBrowserWidth     = "browser_width"
	StoreKeyBrowserHeight    = "browser_height"
)

// LoadConfig loads all browser configuration from the store
func (b *DevBrowser) LoadConfig() {
	// Load auto-start setting
	if val, err := b.db.Get(StoreKeyBrowserAutostart); err == nil && val != "" {
		// Handle legacy values: convert "true"/"false" to "t"/"f"
		switch val {
		case "true", "t":
			b.autoStart = true
			// Migrate legacy value if needed
			if val == "true" {
				b.db.Set(StoreKeyBrowserAutostart, "t")
			}
		case "false", "f":
			b.autoStart = false
			// Migrate legacy value if needed
			if val == "false" {
				b.db.Set(StoreKeyBrowserAutostart, "f")
			}
		default:
			b.autoStart = true // Default if unknown value
		}
	} else {
		b.autoStart = true // Default: auto-start enabled
	}

	// Load position
	if pos, err := b.db.Get(StoreKeyBrowserPosition); err == nil && pos != "" {
		b.position = pos
	}

	// Load width
	if w, err := b.db.Get(StoreKeyBrowserWidth); err == nil && w != "" {
		if width, err := strconv.Atoi(w); err == nil {
			b.width = width
		}
	}

	// Load height
	if h, err := b.db.Get(StoreKeyBrowserHeight); err == nil && h != "" {
		if height, err := strconv.Atoi(h); err == nil {
			b.height = height
		}
	}
}

// SaveConfig saves all browser configuration to the store
func (b *DevBrowser) SaveConfig() error {
	// Save auto-start
	val := "f"
	if b.autoStart {
		val = "t"
	}
	if err := b.db.Set(StoreKeyBrowserAutostart, val); err != nil {
		return err
	}

	// Save position
	if err := b.db.Set(StoreKeyBrowserPosition, b.position); err != nil {
		return err
	}

	// Save width
	if err := b.db.Set(StoreKeyBrowserWidth, strconv.Itoa(b.width)); err != nil {
		return err
	}

	// Save height
	if err := b.db.Set(StoreKeyBrowserHeight, strconv.Itoa(b.height)); err != nil {
		return err
	}

	return nil
}
