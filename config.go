package devbrowser

import (
	"fmt"
)

// Store keys for browser configuration
const (
	StoreKeyBrowserAutostart = "browser_autostart"
	StoreKeyBrowserPosition  = "browser_position"
	StoreKeyBrowserSize      = "browser_size"
	StoreKeyViewportMode     = "viewport_mode"
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

	// Load size (width,height)
	if size, err := b.db.Get(StoreKeyBrowserSize); err == nil && size != "" {
		var w, h int
		if _, err := fmt.Sscanf(size, "%d,%d", &w, &h); err == nil {
			b.width = w
			b.height = h
			b.sizeConfigured = true
		}
	}

	// Load viewport mode
	if mode, err := b.db.Get(StoreKeyViewportMode); err == nil && mode != "" {
		b.viewportMode = mode
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

	// Save size
	size := fmt.Sprintf("%d,%d", b.width, b.height)
	if err := b.db.Set(StoreKeyBrowserSize, size); err != nil {
		return err
	}

	// Save viewport mode
	if err := b.db.Set(StoreKeyViewportMode, b.viewportMode); err != nil {
		return err
	}

	return nil
}
