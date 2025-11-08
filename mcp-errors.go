package devbrowser

import (
	"fmt"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func (b *DevBrowser) getErrorTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_get_errors",
			Description: "Get JavaScript runtime errors and uncaught exceptions to quickly identify crashes, bugs, or WASM panics. Returns error messages with stack traces.",
			Parameters: []ParameterMetadata{
				{
					Name:        "limit",
					Description: "Maximum number of recent errors to return",
					Required:    false,
					Type:        "number",
					Default:     20,
				},
			},
			Execute: func(args map[string]any, progress chan<- string) {
				if !b.isOpen {
					progress <- "Browser is not open. Please open it first with browser_open"
					return
				}

				limit := 20
				if l, ok := args["limit"].(float64); ok {
					limit = int(l)
				}

				b.errorsMutex.Lock()
				defer b.errorsMutex.Unlock()

				if len(b.jsErrors) == 0 {
					progress <- "No JavaScript errors captured"
					return
				}

				start := 0
				if len(b.jsErrors) > limit {
					start = len(b.jsErrors) - limit
				}

				for _, err := range b.jsErrors[start:] {
					progress <- fmt.Sprintf("Error: %s\n  at %s:%d:%d\n%s\n---", err.Message, err.Source, err.LineNumber, err.ColumnNumber, err.StackTrace)
				}
			},
		},
	}
}

func (b *DevBrowser) initializeErrorCapture() {
	chromedp.ListenTarget(b.ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventExceptionThrown:
			b.errorsMutex.Lock()
			defer b.errorsMutex.Unlock()

			exception := ev.ExceptionDetails
			jsErr := JSError{
				Message:      exception.Text,
				Source:       exception.URL,
				LineNumber:   int(exception.LineNumber),
				ColumnNumber: int(exception.ColumnNumber),
				Timestamp:    time.Now(),
			}

			if exception.Exception != nil && exception.Exception.Description != "" {
				jsErr.StackTrace = exception.Exception.Description
			}

			b.jsErrors = append(b.jsErrors, jsErr)
		case *runtime.EventExecutionContextsCleared:
			b.errorsMutex.Lock()
			b.jsErrors = []JSError{}
			b.errorsMutex.Unlock()
		}
	})
}
