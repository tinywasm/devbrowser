package devbrowser

type EmulateDeviceArgs struct {
	Mode     string `input:"-"`
	Capture  bool   `input:"-"`
	Selector string `input:"-"`
}

type GetConsoleArgs struct {
	Lines int `input:"number"`
}

type NavigateArgs struct {
	URL string `input:"-" help:"Absolute URL, or a relative path to the running app, e.g. /login"`
}

type ScreenshotArgs struct {
	Fullpage bool `input:"-"`
}

type InspectElementArgs struct {
	Selector string `input:"-"`
}

type ClickElementArgs struct {
	Selector  string `input:"-"`
	WaitAfter int    `input:"number"`
	Timeout   int    `input:"number"`
}

type FillElementArgs struct {
	Selector  string `input:"-"`
	Value     string `input:"-"`
	WaitAfter int    `input:"number"`
	Timeout   int    `input:"number"`
}

type SwipeElementArgs struct {
	Selector  string `input:"-"`
	Direction string `input:"-"`
	Distance  int    `input:"number"`
}

type EvaluateJSArgs struct {
	Script       string `input:"-"`
	AwaitPromise bool   `input:"-"`
}

type GetNetworkLogsArgs struct {
	Filter string `input:"-"`
	Limit  int    `input:"number"`
}

type GetErrorsArgs struct {
	Limit int `input:"number"`
}

type GetPerformanceArgs struct {
	Reserved int `input:"-"`
}

type GetContentArgs struct {
	Reserved int `input:"-"`
}

type EmptyArgs struct {
	Reserved int `input:"-"`
}

type OpenBrowserArgs struct {
	Port  string `input:"-"`
	Https bool   `input:"-"`
}

type CloseBrowserArgs struct {
	Reserved int `input:"-"`
}

type GetSourceArgs struct {
	Selector string `input:"-"` // vacío = página completa
}

type GetStylesArgs struct {
	Selector string `input:"-"`      // vacío = todas las reglas de todos los stylesheets
	Sheet    int    `input:"number"` // índice de stylesheet (-1 = todos)
}

type GetStorageArgs struct {
	Type string `input:"-"` // "local" | "session" | "cookies" (default: "local")
}

type GetAssetArgs struct {
	URL string `input:"-"` // URL absoluta del archivo JS o CSS a descargar
}

type InterceptRequestArgs struct {
	Action string `input:"-"` // "start" | "stop" | "get"
	Filter string `input:"-"` // filtro de URL (substring), vacío = todo
	Limit  int    `input:"number"`
}

type InterceptedRequest struct {
	URL          string
	Method       string
	RequestBody  string
	ResponseBody string
	Status       int
}
