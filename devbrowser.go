package devbrowser

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/chromedp"
	"golang.design/x/clipboard"
)

type store interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type DevBrowser struct {
	config    serverConfig
	ui        userInterface
	width     int    // ej "800" default "1024"
	height    int    //ej: "600" default "768"
	position  string //ej: "1930,0" (when you have second monitor) default: "0,0"
	headless  bool   // true para modo headless (sin UI), false muestra el navegador
	autoStart bool   // true if browser should auto-open on startup

	isOpen bool // Indica si el navegador está abierto

	db store // Key-value store para configuración y estado

	// chromedp fields
	ctx    context.Context
	cancel context.CancelFunc

	readyChan chan bool
	errChan   chan error
	exitChan  chan bool

	log func(message ...any) // For logging output (Loggable interface)

	// Console log capture
	consoleLogs []string
	logsMutex   sync.Mutex

	// Network log capture
	networkLogs  []NetworkLogEntry
	networkMutex sync.Mutex

	// JS error capture
	jsErrors    []JSError
	errorsMutex sync.Mutex

	// Operation busy flag (atomic) to prevent race conditions and UI blocking
	// 0 = idle, 1 = busy
	busy int32
}

type JSError struct {
	Message      string
	Source       string // File/URL where error occurred
	LineNumber   int
	ColumnNumber int
	StackTrace   string
	Timestamp    time.Time
}

type NetworkLogEntry struct {
	URL       string
	Method    string
	Status    int
	Type      string // xhr, fetch, document, script, image, etc.
	Duration  int64  // milliseconds
	Failed    bool
	ErrorText string
}

type serverConfig interface {
	ServerPort() string
}

type userInterface interface {
	RefreshUI()
	ReturnFocus() error
}

/*
devbrowser.New creates a new DevBrowser instance.

	type serverConfig interface {
		GetServerPort() string
	}

	type userInterface interface {
		RefreshUI()
		ReturnFocus() error
	}

	example :  New(serverConfig, userInterface, exitChan)
*/
func New(sc serverConfig, ui userInterface, st store, exitChan chan bool) *DevBrowser {

	// Initialize clipboard for cross-platform support
	err := clipboard.Init()
	if err != nil {
		// Can't log yet, no logger injected
	}

	browser := &DevBrowser{
		config:    sc,
		ui:        ui,
		db:        st,
		width:     1024,  // Default width
		height:    768,   // Default height
		position:  "0,0", // Default position
		readyChan: make(chan bool),
		errChan:   make(chan error),
		exitChan:  exitChan,
	}

	// Load all configuration from store
	browser.LoadConfig()

	return browser
}

// AutoStart opens the browser if auto-start is enabled in config
// Should be called after the server is ready
// NOTE: OpenBrowser() contains a blocking select, so it runs in a goroutine
func (b *DevBrowser) AutoStart() {
	if !b.autoStart {
		return
	}

	go func() {
		if !atomic.CompareAndSwapInt32(&b.busy, 0, 1) {
			// Already busy
			return
		}
		defer atomic.StoreInt32(&b.busy, 0)

		if !b.isOpen {
			b.OpenBrowser()
		}
	}()
}

func (h *DevBrowser) BrowserStartUrlChanged(fieldName string, oldValue, newValue string) error {

	if !h.isOpen {
		return nil
	}

	return h.RestartBrowser()
}

func (h *DevBrowser) RestartBrowser() error {

	this := errors.New("RestartBrowser")

	err := h.CloseBrowser()
	if err != nil {
		return errors.Join(this, err)
	}

	h.OpenBrowser()

	return nil
}

func (b *DevBrowser) navigateToURL(url string) error {
	if b.ctx == nil {
		return errors.New("context not initialized")
	}

	if err := chromedp.Run(b.ctx, chromedp.Navigate(url)); err != nil {
		return err
	}
	return nil
}

func (b *DevBrowser) Reload() error {
	if b.ctx != nil && b.isOpen {
		b.Logger("Reload")
		if err := chromedp.Run(b.ctx, chromedp.Reload()); err != nil {
			return errors.New("Reload " + err.Error())
		}
	}
	return nil
}

func (b *DevBrowser) SetLog(f func(message ...any)) {
	b.log = f
}

func (b *DevBrowser) Logger(messages ...any) {
	if b.log != nil {
		b.log(messages...)
	}
}

// SetHeadless configura si el navegador debe ejecutarse en modo headless (sin UI).
// Por defecto es false (muestra la ventana del navegador).
// Debe llamarse antes de OpenBrowser().
func (b *DevBrowser) SetHeadless(headless bool) {
	b.headless = headless
}

// monitorBrowserClose monitors the browser context and updates state when browser is closed manually
func (b *DevBrowser) monitorBrowserClose() {
	if b.ctx == nil {
		return
	}

	// Wait for context to be done (browser closed)
	<-b.ctx.Done()

	// Only handle if browser was marked as open (manual close by user)
	if b.isOpen {
		b.Logger("Browser closed by user")
		b.isOpen = false
		b.ctx = nil
		b.cancel = nil
		b.ui.RefreshUI()
	}
}
