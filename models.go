package devbrowser

import "github.com/tinywasm/model"

var ScreenshotArgsModel = model.Definition{
	Name: "screenshot_args",
	Fields: model.Fields{
		{Name: "fullpage", Type: model.Bool()},
	},
}

var ClickElementArgsModel = model.Definition{
	Name: "click_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.Text(), NotNull: true},
		{Name: "wait_after", Type: model.Int()},
		{Name: "timeout", Type: model.Int()},
	},
}

var NavigateArgsModel = model.Definition{
	Name: "navigate_args",
	Fields: model.Fields{
		{Name: "url", Type: model.Text(), NotNull: true},
	},
}

var EmulateDeviceArgsModel = model.Definition{
	Name: "emulate_device_args",
	Fields: model.Fields{
		{Name: "mode", Type: model.Text()},
		{Name: "capture", Type: model.Bool()},
		{Name: "selector", Type: model.Text()},
	},
}

var GetConsoleArgsModel = model.Definition{
	Name: "get_console_args",
	Fields: model.Fields{
		{Name: "lines", Type: model.Int()},
	},
}

var FillElementArgsModel = model.Definition{
	Name: "fill_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.Text(), NotNull: true},
		{Name: "value", Type: model.Text(), NotNull: true},
		{Name: "wait_after", Type: model.Int()},
		{Name: "timeout", Type: model.Int()},
	},
}

var SwipeElementArgsModel = model.Definition{
	Name: "swipe_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.Text(), NotNull: true},
		{Name: "direction", Type: model.Text(), NotNull: true},
		{Name: "distance", Type: model.Int()},
	},
}

var EvaluateJSArgsModel = model.Definition{
	Name: "evaluate_js_args",
	Fields: model.Fields{
		{Name: "script", Type: model.Text(), NotNull: true},
		{Name: "await_promise", Type: model.Bool()},
	},
}

var GetNetworkLogsArgsModel = model.Definition{
	Name: "get_network_logs_args",
	Fields: model.Fields{
		{Name: "filter", Type: model.Text()},
		{Name: "limit", Type: model.Int()},
	},
}

var GetErrorsArgsModel = model.Definition{
	Name: "get_errors_args",
	Fields: model.Fields{
		{Name: "limit", Type: model.Int()},
	},
}

var GetPerformanceArgsModel = model.Definition{
	Name: "get_performance_args",
	Fields: model.Fields{},
}

var GetContentArgsModel = model.Definition{
	Name: "get_content_args",
	Fields: model.Fields{},
}

var GetSourceArgsModel = model.Definition{
	Name: "get_source_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.Text()},
	},
}

var InspectElementArgsModel = model.Definition{
	Name: "inspect_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.Text(), NotNull: true},
	},
}

var GetStylesArgsModel = model.Definition{
	Name: "get_styles_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.Text()},
		{Name: "sheet", Type: model.Int()},
	},
}

var GetStorageArgsModel = model.Definition{
	Name: "get_storage_args",
	Fields: model.Fields{
		{Name: "type", Type: model.Text()},
	},
}

var GetAssetArgsModel = model.Definition{
	Name: "get_asset_args",
	Fields: model.Fields{
		{Name: "url", Type: model.Text(), NotNull: true},
	},
}

var InterceptRequestArgsModel = model.Definition{
	Name: "intercept_request_args",
	Fields: model.Fields{
		{Name: "action", Type: model.Text(), NotNull: true},
		{Name: "filter", Type: model.Text()},
		{Name: "limit", Type: model.Int()},
	},
}

var OpenBrowserArgsModel = model.Definition{
	Name: "open_browser_args",
	Fields: model.Fields{
		{Name: "port", Type: model.Text()},
		{Name: "https", Type: model.Bool()},
	},
}

var CloseBrowserArgsModel = model.Definition{
	Name: "close_browser_args",
	Fields: model.Fields{},
}

type InterceptedRequest struct {
	URL          string
	Method       string
	RequestBody  string
	ResponseBody string
	Status       int
}
