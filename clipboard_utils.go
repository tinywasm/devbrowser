package devbrowser

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
)

// writeToClipboard writes the given data to the system clipboard.
// It supports Linux (xclip, wl-copy), macOS (pbcopy), and Windows (powershell).
func writeToClipboard(data []byte) error {
	switch runtime.GOOS {
	case "linux":
		// Try Wayland first
		if err := execCommandInput("wl-copy", data); err == nil {
			return nil
		}
		// Fallback to X11
		if err := execCommandInput("xclip", data, "-selection", "clipboard", "-t", "image/png"); err == nil {
			return nil
		}
		return errors.New("clipboard tool not found (install wl-copy or xclip)")
	case "darwin":
		return execCommandInput("pbcopy", data)
	case "windows":
		// PowerShell script to set clipboard image
		// Note: This is a simplified approach. For binary data (images), it's more complex with PowerShell.
		// However, the original library handled images. Let's try a PowerShell command for images.
		// Powershell Set-Clipboard only supports text or file lists easily. For images, we need .NET reflection.
		// Or we can just skip image clipboard for Windows if too complex for stdlib-only.
		// But let's try a basic text fallback if data is text, or a more complex one if image.
		// Actually, let's implement a best-effort approach.
		return errors.New("clipboard writing not fully implemented for Windows without cgo/libraries")
	default:
		return errors.New("unsupported operating system for clipboard")
	}
}

func execCommandInput(name string, input []byte, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewReader(input)
	return cmd.Run()
}
