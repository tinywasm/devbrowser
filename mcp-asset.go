package devbrowser

import (
	"fmt"

	"github.com/tinywasm/context"
	"github.com/tinywasm/devbrowser/cdproto/runtime"
	"github.com/tinywasm/devbrowser/chromedp"
	"github.com/tinywasm/mcp"
)

func (b *DevBrowser) GetAssetTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "browser_get_asset",
			Description: "Download the content of a JS or CSS file by Url using the active session. Evades CORS and authentication issues by fetching from the browser context.",
			Args: new(GetAssetArgs),
			Resource:    "browser",
			Action:      'r',
			Execute: func(ctx *context.Context, req mcp.Request) (*mcp.Result, error) {
				if !b.IsOpenFlag {
					return nil, fmt.Errorf("Browser is not open. Please open it first with browser_open")
				}
				var args GetAssetArgs
				if err := req.Bind(&args); err != nil {
					return nil, err
				}

				js := fmt.Sprintf(`fetch(%q).then(r => {
					if (!r.ok) throw new Error("HTTP error! status: " + r.status);
					return r.text();
				})`, args.Url)

				var result interface{}
				err := chromedp.Run(b.Ctx,
					chromedp.Evaluate(js, &result, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
						return p.WithAwaitPromise(true)
					}),
				)
				if err != nil {
					return nil, err
				}

				resStr, ok := result.(string)
				if !ok {
					return nil, fmt.Errorf("Expected string result from asset fetch, got %T", result)
				}

				return mcp.Text(resStr), nil
			},
		},
	}
}
