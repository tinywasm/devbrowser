package devbrowser

import (
	"errors"

	"github.com/chromedp/chromedp"
)

// GetConsoleLogs captures and returns console logs from the browser.
// It executes JavaScript to retrieve console messages and returns them as a slice of strings.
// Returns an error if the browser context is not initialized or if there's an error executing the script.
func (b *DevBrowser) GetConsoleLogs() ([]string, error) {
	if b.ctx == nil {
		return nil, errors.New("browser context not initialized")
	}

	var logs []string

	// JavaScript code to retrieve console logs
	// This captures console.log, console.error, console.warn, and console.info
	script := `
		(function() {
			if (!window.__consoleLogs) {
				window.__consoleLogs = [];
				
				['log', 'error', 'warn', 'info'].forEach(function(method) {
					var original = console[method];
					console[method] = function() {
						var args = Array.prototype.slice.call(arguments);
						var message = args.map(function(arg) {
							if (typeof arg === 'object') {
								try {
									return JSON.stringify(arg);
								} catch(e) {
									return String(arg);
								}
							}
							return String(arg);
						}).join(' ');
						
						window.__consoleLogs.push('[' + method.toUpperCase() + '] ' + message);
						original.apply(console, arguments);
					};
				});
			}
			return window.__consoleLogs;
		})();
	`

	err := chromedp.Run(b.ctx,
		chromedp.Evaluate(script, &logs),
	)

	if err != nil {
		return nil, errors.New("GetConsoleLogs: " + err.Error())
	}

	return logs, nil
}

// ClearConsoleLogs clears the captured console logs in the browser.
func (b *DevBrowser) ClearConsoleLogs() error {
	if b.ctx == nil {
		return errors.New("browser context not initialized")
	}

	script := `
		if (window.__consoleLogs) {
			window.__consoleLogs = [];
		}
	`

	err := chromedp.Run(b.ctx,
		chromedp.Evaluate(script, nil),
	)

	if err != nil {
		return errors.New("ClearConsoleLogs: " + err.Error())
	}

	return nil
}
