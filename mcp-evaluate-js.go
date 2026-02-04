package devbrowser

import "github.com/tinywasm/mcpserve"

import (
	"encoding/json"
	"fmt"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func (b *DevBrowser) getEvaluateJsTools() []mcpserve.ToolMetadata {
	return []mcpserve.ToolMetadata{
		{
			Name:        "browser_evaluate_js",
			Description: "Execute JavaScript code in browser context to inspect DOM, call WASM exports, test functions, or debug application state. Returns execution result or error.",
			Parameters: []mcpserve.ParameterMetadata{
				{
					Name:        "script",
					Description: "JavaScript code to execute in browser context",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "await_promise",
					Description: "Wait for Promise resolution if script returns Promise",
					Required:    false,
					Type:        "boolean",
					Default:     false,
				},
			},
			Execute: func(args map[string]any) {
				if !b.isOpen {
					b.Logger("Browser is not open. Please open it first with browser_open")
					return
				}

				script, ok := args["script"].(string)
				if !ok || script == "" {
					b.Logger("Script parameter is required")
					return
				}

				awaitPromise, _ := args["await_promise"].(bool)

				var res interface{}
				err := chromedp.Run(b.ctx,
					chromedp.Evaluate(script, &res, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
						return p.WithAwaitPromise(awaitPromise)
					}),
				)

				if err != nil {
					b.Logger(fmt.Sprintf("Error: %v", err))
					return
				}

				switch v := res.(type) {
				case nil:
					b.Logger("undefined")
				case string:
					b.Logger(v)
				case float64, bool:
					b.Logger(fmt.Sprintf("%v", v))
				default:
					jsonRes, err := json.Marshal(v)
					if err != nil {
						b.Logger(fmt.Sprintf("Error: failed to serialize result: %v", err))
						return
					}
					b.Logger(string(jsonRes))
				}
			},
		},
	}
}
