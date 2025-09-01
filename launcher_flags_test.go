package devbrowser

import (
	"testing"

	"github.com/go-rod/rod/lib/launcher"
)

func TestLauncherAppendVariousFlags(t *testing.T) {
	// test a representative set of flags

	l := launcher.New()
	panicked := false
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			t.Logf("launcher.Append panicked: %v", r)
		}
	}()

	// Append many flags as untyped string constants to match expected usage
	l.Append("--disable-blink-features=WebFontsInterventionV2")
	l.Append("--use-fake-ui-for-media-stream")
	l.Append("--no-focus-on-load")
	l.Append("--auto-open-devtools-for-tabs")
	l.Append("--window-position=0,0")

	if panicked {
		t.Fatalf("launcher.Append panicked unexpectedly")
	} else {
		t.Log("launcher.Append completed without panic for tested flags")
	}
}
