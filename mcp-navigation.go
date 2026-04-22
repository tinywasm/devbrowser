package devbrowser

import (
	"fmt"

	"github.com/tinywasm/context"
	"github.com/tinywasm/mcp"
)

func (b *DevBrowser) GetNavigationTools() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "browser_navigate",
			Description: "Navigate the browser to a specific URL",
			InputSchema: EncodeSchema(new(NavigateArgs)),
			Resource:    "browser",
			Action:      'u',
			Execute: func(Ctx *context.Context, req mcp.Request) (*mcp.Result, error) {
				if !b.IsOpenFlag {
					return nil, fmt.Errorf("Browser is not open. Please open it first with browser_open")
				}

				var args NavigateArgs
				if err := req.Bind(&args); err != nil {
					return nil, err
				}

				err := b.NavigateToURL(args.URL)
				if err != nil {
					return nil, fmt.Errorf("Error navigating to %s: %v", args.URL, err)
				}

				return mcp.Text(fmt.Sprintf("Navigated to %s", args.URL)), nil
			},
		},
	}
}
