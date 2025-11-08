package devbrowser

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func (b *DevBrowser) getNetworkTools() []ToolMetadata {
	return []ToolMetadata{
		{
			Name:        "browser_get_network_logs",
			Description: "Get network requests and responses to debug API calls, asset loading failures, CORS errors, or slow requests. Shows URL, status, method, and timing.",
			Parameters: []ParameterMetadata{
				{
					Name:        "filter",
					Description: "Filter by request type",
					Required:    false,
					Type:        "string",
					EnumValues:  []string{"all", "xhr", "fetch", "document", "script", "image"},
					Default:     "all",
				},
				{
					Name:        "limit",
					Description: "Maximum number of recent requests to return",
					Required:    false,
					Type:        "number",
					Default:     50,
				},
			},
			Execute: func(args map[string]any, progress chan<- string) {
				if !b.isOpen {
					progress <- "Browser is not open. Please open it first with browser_open"
					return
				}

				filter := "all"
				if f, ok := args["filter"].(string); ok {
					filter = f
				}

				limit := 50
				if l, ok := args["limit"].(float64); ok {
					limit = int(l)
				}

				b.networkMutex.Lock()
				defer b.networkMutex.Unlock()

				var filteredLogs []NetworkLogEntry
				for _, log := range b.networkLogs {
					if filter == "all" || strings.ToLower(log.Type) == filter {
						filteredLogs = append(filteredLogs, log)
					}
				}

				if len(filteredLogs) > limit {
					filteredLogs = filteredLogs[len(filteredLogs)-limit:]
				}

				if len(filteredLogs) == 0 {
					if filter == "all" {
						progress <- "No network requests captured"
					} else {
						progress <- fmt.Sprintf("No %s requests found", filter)
					}
					return
				}

				for _, log := range filteredLogs {
					status := fmt.Sprintf("%d", log.Status)
					if log.Failed {
						status = "Failed"
					}
					progress <- fmt.Sprintf("%s %s %s (%dms) [%s] %s", status, log.Method, log.URL, log.Duration, log.Type, log.ErrorText)
				}
			},
		},
	}
}

func (b *DevBrowser) initializeNetworkCapture() {
	type requestInfo struct {
		Method string
		URL    string
		Time   time.Time
	}
	requests := make(map[network.RequestID]requestInfo)
	var mutex sync.Mutex

	chromedp.ListenTarget(b.ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			mutex.Lock()
			requests[ev.RequestID] = requestInfo{
				Method: ev.Request.Method,
				URL:    ev.Request.URL,
				Time:   time.Now(),
			}
			mutex.Unlock()
			if ev.Type == "Document" {
				b.networkMutex.Lock()
				b.networkLogs = []NetworkLogEntry{}
				b.networkMutex.Unlock()
			}

		case *network.EventResponseReceived:
			mutex.Lock()
			reqInfo, ok := requests[ev.RequestID]
			mutex.Unlock()
			if ok {
				duration := time.Since(reqInfo.Time).Milliseconds()
				b.networkMutex.Lock()
				b.networkLogs = append(b.networkLogs, NetworkLogEntry{
					URL:      ev.Response.URL,
					Method:   reqInfo.Method,
					Status:   int(ev.Response.Status),
					Type:     string(ev.Type),
					Duration: duration,
				})
				b.networkMutex.Unlock()
			}

		case *network.EventLoadingFailed:
			mutex.Lock()
			reqInfo, ok := requests[ev.RequestID]
			mutex.Unlock()
			if ok {
				duration := time.Since(reqInfo.Time).Milliseconds()
				b.networkMutex.Lock()
				b.networkLogs = append(b.networkLogs, NetworkLogEntry{
					URL:       reqInfo.URL,
					Method:    reqInfo.Method,
					Type:      string(ev.Type),
					Duration:  duration,
					Failed:    true,
					ErrorText: ev.ErrorText,
				})
				b.networkMutex.Unlock()
			}
		}
	})
}
