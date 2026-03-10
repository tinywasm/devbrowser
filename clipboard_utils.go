package devbrowser

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"runtime"
)

// writeToClipboard writes the given data to the system clipboard.
// It supports Linux (wl-copy on Wayland, xclip on X11), macOS (pbcopy), and Windows (powershell).
func writeToClipboard(data []byte) error {
	switch runtime.GOOS {
	case "linux":
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			if err := execCommandInput("wl-copy", data); err == nil {
				return nil
			}
			return errors.New("clipboard tool not found: run `sudo apt install wl-clipboard`")
		}
		if err := execCommandInput("xclip", data, "-selection", "clipboard", "-t", "image/png"); err == nil {
			return nil
		}
		return errors.New("clipboard tool not found: run `sudo apt install xclip`")
	case "darwin":
		return execCommandInput("pbcopy", data)
	case "windows":
		return errors.New("clipboard writing not supported on Windows without cgo/libraries")
	default:
		return errors.New("unsupported operating system for clipboard")
	}
}

func execCommandInput(name string, input []byte, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewReader(input)
	return cmd.Run()
}
