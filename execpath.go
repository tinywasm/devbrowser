package devbrowser

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// ResolveChromeExecPath finds a working Chrome/Chromium executable.
// It prioritizes Google Chrome over Chromium to avoid known regressions in some
// Chromium packages (e.g., Debian's chromium 150+).
// It also validates each candidate by running it with --version.
func ResolveChromeExecPath() string {
	// 1. Honor explicit environment variable override
	if envPath := os.Getenv("CHROME_EXECPATH"); envPath != "" {
		if validateExecPath(envPath) {
			return envPath
		}
	}

	// 2. Build candidate list based on OS
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
		}
	case "windows":
		candidates = []string{
			"chrome.exe",
			"chrome",
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			os.ExpandEnv(`${USERPROFILE}\AppData\Local\Google\Chrome\Application\chrome.exe`),
			os.ExpandEnv(`${USERPROFILE}\AppData\Local\Chromium\Application\chrome.exe`),
		}
	default:
		// Unix-like: prioritize Google Chrome over Chromium
		candidates = []string{
			"google-chrome",
			"google-chrome-stable",
			"google-chrome-beta",
			"google-chrome-unstable",
			"chromium",
			"chromium-browser",
			"headless-shell",
			"headless_shell",
			"/usr/bin/google-chrome",
			"/usr/local/bin/chrome",
			"/snap/bin/chromium",
			"chrome",
		}
	}

	for _, name := range candidates {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if validateExecPath(path) {
			return path
		}
	}

	// Fallback to "google-chrome" if nothing is found, to let chromedp try its default
	// and produce its own error if it still fails.
	return "google-chrome"
}

// validateExecPath runs the executable with --version and a timeout to ensure it works.
func validateExecPath(path string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
