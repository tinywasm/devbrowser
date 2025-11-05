package devbrowser

// Utilities for tests: provide a single initializer that tests can call to
// create a trimmed-down DevBrowser suitable for unit tests. The function
// accepts variadic options so callers don't have to pass anything. If a
// logger function or a custom exit channel is provided it will be used.
//
// Usage examples:
//   db, exit := DefaultTestBrowser()
//   db, exit := DefaultTestBrowser(myLogger)
//   db, exit := DefaultTestBrowser(myLogger, myExitChan)

// Note: we provide simple local implementations of the required
// interfaces so tests don't need to construct or import fake types.

// defaultServerConfig implements serverConfig for tests.
type defaultServerConfig struct{}

func (defaultServerConfig) ServerPort() string { return "0" }

// defaultUI implements userInterface for tests.
type defaultUI struct{}

func (defaultUI) RefreshUI() {}

func (defaultUI) ReturnFocus() error { return nil }

// DefaultTestBrowser creates a DevBrowser instance for tests.
//
// Accepted variadic options (in any order):
// - func(...any) : a logger function. If omitted a no-op logger is used.
// - chan bool     : a custom exit channel. If omitted a new channel is created.
// Any other option types are ignored.
func DefaultTestBrowser(opts ...any) (*DevBrowser, chan bool) {
	var logger func(message ...any)
	var exit chan bool

	for _, o := range opts {
		switch v := o.(type) {
		case func(...any):
			logger = v
		case chan bool:
			exit = v
		}
	}

	if logger == nil {
		logger = func(...any) {}
	}
	if exit == nil {
		exit = make(chan bool)
	}

	db := New(defaultServerConfig{}, defaultUI{}, exit, logger)
	db.SetHeadless(true) // Los tests usan modo headless por defecto
	return db, exit
}
