package devbrowser

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tinywasm/context"
	"github.com/tinywasm/devbrowser/cdproto/network"
	"github.com/tinywasm/devbrowser/chromedp"
	"github.com/tinywasm/mcp"
)

func (b *DevBrowser) GetNetworkTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "browser_get_network_logs",
			Description: "Get network requests and responses to debug API calls, asset loading failures, CORS errors, or slow requests. Shows URL, status, method, and timing.",
			InputSchema: EncodeSchema(new(GetNetworkLogsArgs)),
			Resource:    "browser",
			Action:      'r',
			Execute: func(Ctx *context.Context, req mcp.Request) (*mcp.Result, error) {
				if !b.IsOpenFlag {
					return nil, fmt.Errorf("Browser is not open. Please open it first with browser_open")
				}

				var args GetNetworkLogsArgs
				if err := req.Bind(&args); err != nil {
					return nil, err
				}

				filter := args.Filter
				if filter == "" {
					filter = "all"
				}

				limit := args.Limit
				if limit == 0 {
					limit = 50
				}

				b.NetworkMutex.Lock()
				defer b.NetworkMutex.Unlock()

				var filteredLogs []NetworkLogEntry
				for _, log := range b.NetworkLogs {
					if filter == "all" || strings.ToLower(log.Type) == filter {
						filteredLogs = append(filteredLogs, log)
					}
				}

				if len(filteredLogs) > limit {
					filteredLogs = filteredLogs[len(filteredLogs)-limit:]
				}

				if len(filteredLogs) == 0 {
					var msg string
					if filter == "all" {
						msg = "No network requests captured"
					} else {
						msg = fmt.Sprintf("No %s requests found", filter)
					}
					return mcp.Text(msg), nil
				}

				var result strings.Builder
				for i, log := range filteredLogs {
					if i > 0 {
						result.WriteString("\n")
					}
					status := fmt.Sprintf("%d", log.Status)
					if log.Failed {
						status = "Failed"
					}
					result.WriteString(fmt.Sprintf("%s %s %s (%dms) [%s] %s", status, log.Method, log.URL, log.Duration, log.Type, log.ErrorText))
				}

				return mcp.Text(result.String()), nil
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

	chromedp.ListenTarget(b.Ctx, func(ev interface{}) {
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
				b.NetworkMutex.Lock()
				b.NetworkLogs = []NetworkLogEntry{}
				b.NetworkMutex.Unlock()
			}

		case *network.EventResponseReceived:
			mutex.Lock()
			reqInfo, ok := requests[ev.RequestID]
			mutex.Unlock()
			if ok {
				duration := time.Since(reqInfo.Time).Milliseconds()
				b.NetworkMutex.Lock()
				b.NetworkLogs = append(b.NetworkLogs, NetworkLogEntry{
					URL:      ev.Response.URL,
					Method:   reqInfo.Method,
					Status:   int(ev.Response.Status),
					Type:     string(ev.Type),
					Duration: duration,
				})
				b.NetworkMutex.Unlock()
			}

		case *network.EventLoadingFailed:
			mutex.Lock()
			reqInfo, ok := requests[ev.RequestID]
			mutex.Unlock()
			if ok {
				duration := time.Since(reqInfo.Time).Milliseconds()
				b.NetworkMutex.Lock()
				b.NetworkLogs = append(b.NetworkLogs, NetworkLogEntry{
					URL:       reqInfo.URL,
					Method:    reqInfo.Method,
					Type:      string(ev.Type),
					Duration:  duration,
					Failed:    true,
					ErrorText: ev.ErrorText,
				})
				b.NetworkMutex.Unlock()
			}
		}
	})
}
