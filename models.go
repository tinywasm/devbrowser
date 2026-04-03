package devbrowser

// ormc:formonly
type EmulateDeviceArgs struct {
	Mode     string `db:"not_null" input:"-"`
	Capture  bool   `input:"-"`
	Selector string `input:"-"`
}

// ormc:formonly
type GetConsoleArgs struct {
	Lines int `input:"number"`
}

// ormc:formonly
type NavigateArgs struct {
	URL string `db:"not_null" input:"-"`
}

// ormc:formonly
type ScreenshotArgs struct {
	Fullpage bool `input:"-"`
}

// ormc:formonly
type InspectElementArgs struct {
	Selector string `db:"not_null" input:"-"`
}

// ormc:formonly
type ClickElementArgs struct {
	Selector  string `db:"not_null" input:"-"`
	WaitAfter int    `input:"number"`
	Timeout   int    `input:"number"`
}

// ormc:formonly
type FillElementArgs struct {
	Selector  string `db:"not_null" input:"-"`
	Value     string `db:"not_null" input:"-"`
	WaitAfter int    `input:"number"`
	Timeout   int    `input:"number"`
}

// ormc:formonly
type SwipeElementArgs struct {
	Selector  string `db:"not_null" input:"-"`
	Direction string `db:"not_null" input:"-"`
	Distance  int    `db:"not_null" input:"number"`
}

// ormc:formonly
type EvaluateJSArgs struct {
	Script       string `db:"not_null" input:"-"`
	AwaitPromise bool   `input:"-"`
}

// ormc:formonly
type GetNetworkLogsArgs struct {
	Filter string `input:"-"`
	Limit  int    `input:"number"`
}

// ormc:formonly
type GetErrorsArgs struct {
	Limit int `input:"number"`
}

// ormc:formonly
type GetPerformanceArgs struct {
	Reserved int `input:"-"`
}

// ormc:formonly
type GetContentArgs struct {
	Reserved int `input:"-"`
}

// ormc:formonly
type EmptyArgs struct {
	Reserved int `input:"-"`
}
