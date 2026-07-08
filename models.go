package devbrowser

import "github.com/tinywasm/model"

var ScreenshotArgsModel = model.Definition{
	Name: "screenshot_args",
	Fields: model.Fields{
		{Name: "fullpage", Type: model.FieldBool},
	},
}

var ClickElementArgsModel = model.Definition{
	Name: "click_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.FieldText, NotNull: true},
		{Name: "wait_after", Type: model.FieldInt},
		{Name: "timeout", Type: model.FieldInt},
	},
}

var NavigateArgsModel = model.Definition{
	Name: "navigate_args",
	Fields: model.Fields{
		{Name: "url", Type: model.FieldText, NotNull: true},
	},
}

var EmulateDeviceArgsModel = model.Definition{
	Name: "emulate_device_args",
	Fields: model.Fields{
		{Name: "mode", Type: model.FieldText},
		{Name: "capture", Type: model.FieldBool},
		{Name: "selector", Type: model.FieldText},
	},
}

var GetConsoleArgsModel = model.Definition{
	Name: "get_console_args",
	Fields: model.Fields{
		{Name: "lines", Type: model.FieldInt},
	},
}

var FillElementArgsModel = model.Definition{
	Name: "fill_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.FieldText, NotNull: true},
		{Name: "value", Type: model.FieldText, NotNull: true},
		{Name: "wait_after", Type: model.FieldInt},
		{Name: "timeout", Type: model.FieldInt},
	},
}

var SwipeElementArgsModel = model.Definition{
	Name: "swipe_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.FieldText, NotNull: true},
		{Name: "direction", Type: model.FieldText, NotNull: true},
		{Name: "distance", Type: model.FieldInt},
	},
}

var EvaluateJSArgsModel = model.Definition{
	Name: "evaluate_js_args",
	Fields: model.Fields{
		{Name: "script", Type: model.FieldText, NotNull: true},
		{Name: "await_promise", Type: model.FieldBool},
	},
}

var GetNetworkLogsArgsModel = model.Definition{
	Name: "get_network_logs_args",
	Fields: model.Fields{
		{Name: "filter", Type: model.FieldText},
		{Name: "limit", Type: model.FieldInt},
	},
}

var GetErrorsArgsModel = model.Definition{
	Name: "get_errors_args",
	Fields: model.Fields{
		{Name: "limit", Type: model.FieldInt},
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
		{Name: "selector", Type: model.FieldText},
	},
}

var InspectElementArgsModel = model.Definition{
	Name: "inspect_element_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.FieldText, NotNull: true},
	},
}

var GetStylesArgsModel = model.Definition{
	Name: "get_styles_args",
	Fields: model.Fields{
		{Name: "selector", Type: model.FieldText},
		{Name: "sheet", Type: model.FieldInt},
	},
}

var GetStorageArgsModel = model.Definition{
	Name: "get_storage_args",
	Fields: model.Fields{
		{Name: "type", Type: model.FieldText},
	},
}

var GetAssetArgsModel = model.Definition{
	Name: "get_asset_args",
	Fields: model.Fields{
		{Name: "url", Type: model.FieldText, NotNull: true},
	},
}

var InterceptRequestArgsModel = model.Definition{
	Name: "intercept_request_args",
	Fields: model.Fields{
		{Name: "action", Type: model.FieldText, NotNull: true},
		{Name: "filter", Type: model.FieldText},
		{Name: "limit", Type: model.FieldInt},
	},
}

var OpenBrowserArgsModel = model.Definition{
	Name: "open_browser_args",
	Fields: model.Fields{
		{Name: "port", Type: model.FieldText},
		{Name: "https", Type: model.FieldBool},
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
