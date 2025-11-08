package devbrowser

import (
	"errors"
	"fmt"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// initializeConsoleCapture sets up the console log capturing system using Chrome DevTools Protocol.
// This captures ALL console messages including those from page load, using runtime events.
func (b *DevBrowser) initializeConsoleCapture() error {
	if b.ctx == nil {
		return errors.New("browser context not initialized")
	}

	// Initialize the console logs slice
	b.consoleLogs = []string{}

	// Listen for console API called events and console cleared events
	chromedp.ListenTarget(b.ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventConsoleAPICalled:
			b.logsMutex.Lock()
			defer b.logsMutex.Unlock()

			// Format the arguments without prefix
			var message string
			for i, arg := range ev.Args {
				if i > 0 {
					message += " "
				}
				// Get the string value from the RemoteObject
				if arg.Value != nil {
					// Extract the raw value without JSON encoding
					val := fmt.Sprintf("%v", arg.Value)
					// Remove surrounding quotes if it's a string value
					if len(val) > 2 && val[0] == '"' && val[len(val)-1] == '"' {
						val = val[1 : len(val)-1]
					}
					message += val
				} else if arg.Description != "" {
					message += arg.Description
				}
			}

			// Add to logs without type prefix to save tokens
			b.consoleLogs = append(b.consoleLogs, message)

		case *runtime.EventExecutionContextsCleared:
			// Clear logs when execution contexts are cleared (page reload/navigation)
			b.logsMutex.Lock()
			b.consoleLogs = []string{}
			b.logsMutex.Unlock()
		}
	})

	// Enable console domain to start receiving events
	err := chromedp.Run(b.ctx, runtime.Enable())
	if err != nil {
		return errors.New("initializeConsoleCapture: " + err.Error())
	}

	return nil
}

// GetConsoleLogs returns captured console logs from the browser.
// Returns an error if the browser context is not initialized.
func (b *DevBrowser) GetConsoleLogs() ([]string, error) {
	if b.ctx == nil {
		return nil, errors.New("browser context not initialized")
	}

	b.logsMutex.Lock()
	defer b.logsMutex.Unlock()

	// Return a copy of the logs to avoid race conditions
	logsCopy := make([]string, len(b.consoleLogs))
	copy(logsCopy, b.consoleLogs)

	return logsCopy, nil
}

// ClearConsoleLogs clears the captured console logs.
func (b *DevBrowser) ClearConsoleLogs() error {
	if b.ctx == nil {
		return errors.New("browser context not initialized")
	}

	b.logsMutex.Lock()
	defer b.logsMutex.Unlock()

	b.consoleLogs = []string{}
	return nil
}
