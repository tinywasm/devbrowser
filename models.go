package devbrowser

import (
	"github.com/tinywasm/mcp"
)

// ormc:formonly
type EmulateDeviceArgs struct {
	Mode     string `input:"required,enum=desktop;mobile;tablet;off"`
	Capture  bool   `input:"default=false"`
	Selector string `input:"optional"`
}

// ormc:formonly
type GetConsoleArgs struct {
	Lines int `input:"default=50"`
}

// ormc:formonly
type NavigateArgs struct {
	URL string `input:"required"`
}

// ormc:formonly
type ScreenshotArgs struct {
	Fullpage bool `input:"default=false"`
}

// ormc:formonly
type InspectElementArgs struct {
	Selector string `input:"required"`
}

// ormc:formonly
type ClickElementArgs struct {
	Selector  string `input:"required"`
	WaitAfter int    `input:"default=100"`
	Timeout   int    `input:"default=5000"`
}

// ormc:formonly
type FillElementArgs struct {
	Selector  string `input:"required"`
	Value     string `input:"required"`
	WaitAfter int    `input:"default=100"`
	Timeout   int    `input:"default=5000"`
}

// ormc:formonly
type SwipeElementArgs struct {
	Selector  string `input:"required"`
	Direction string `input:"required,enum=up;down;left;right"`
	Distance  int    `input:"required"`
}

// ormc:formonly
type EvaluateJSArgs struct {
	Script       string `input:"required"`
	AwaitPromise bool   `input:"default=false"`
}

// ormc:formonly
type EmptyArgs struct{}

func (a *EmulateDeviceArgs) Schema() string   { return "" } // ormc will overwrite
func (a *GetConsoleArgs) Schema() string      { return "" } // ormc will overwrite
func (a *NavigateArgs) Schema() string        { return "" } // ormc will overwrite
func (a *ScreenshotArgs) Schema() string      { return "" } // ormc will overwrite
func (a *InspectElementArgs) Schema() string  { return "" } // ormc will overwrite
func (a *ClickElementArgs) Schema() string    { return "" } // ormc will overwrite
func (a *FillElementArgs) Schema() string     { return "" } // ormc will overwrite
func (a *SwipeElementArgs) Schema() string    { return "" } // ormc will overwrite
func (a *EvaluateJSArgs) Schema() string      { return "" } // ormc will overwrite
func (a *EmptyArgs) Schema() string           { return "" } // ormc will overwrite
