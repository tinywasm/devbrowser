package devbrowser

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tinywasm/devbrowser/chromedp"
	"golang.design/x/clipboard"
)

type store interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type DevBrowser struct {
	ui             userInterface
	width          int    // ej "800" default "1024"
	height         int    //ej: "600" default "768"
	position       string //ej: "1930,0" (when you have second monitor) default: "0,0"
	headless       bool   // true para modo headless (sin UI), false muestra el navegador
	autoStart      bool   // true if browser should auto-open on startup
	monitorWidth   int    // Detected monitor availability width
	monitorHeight  int    // Detected monitor availability height
	sizeConfigured bool   // Track if size was loaded from storage
	viewportMode   string // Current emulation mode ("mobile", "tablet", "desktop", "off", "")
	firstCall      bool   // Internal flag to track if OpenBrowser was called for the first time
	openedOnce     bool   // Internal flag to track if browser was actually opened at least once

	lastPort  string
	lastHttps bool

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

	testMode bool // Skip opening browser in tests

	// Cache configuration
	cacheEnabled bool // Disabled by default for development
	mu           sync.Mutex
}

// Option configures the DevBrowser
type Option func(*DevBrowser)

// WithCache configures whether the browser cache is enabled
func WithCache(enabled bool) Option {
	return func(b *DevBrowser) {
		b.cacheEnabled = enabled
	}
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

	example :  New(userInterface, st, exitChan, WithCache(true))
*/
func New(ui userInterface, st store, exitChan chan bool, opts ...Option) *DevBrowser {

	// Initialize clipboard for cross-platform support
	err := clipboard.Init()
	if err != nil {
		// Can't log yet, no logger injected
	}

	browser := &DevBrowser{
		ui:           ui,
		db:           st,
		width:        1024, // Default width
		height:       768,  // Default height
		position:     "0,0",
		firstCall:    true,
		readyChan:    make(chan bool),
		errChan:      make(chan error),
		exitChan:     exitChan,
		cacheEnabled: false, // Default: Cache disabled for development
	}

	// Apply options
	for _, opt := range opts {
		opt(browser)
	}

	// Load all configuration from store
	browser.LoadConfig()

	//id := atomic.AddInt32(&instanceCounter, 1)
	// Logger not set yet, so we can't log this via b.Logger consistently
	// But let's try just in case it's injected early or for future ref inspection
	// We'll use fmt temporarily for this one-time struct init check if needed,
	// but user dislikes fmt. Let's rely on AutoStart logs primarily.
	// If we really need New log, we might need fmt. But user said no fmt.
	// We'll skip logging in New for now unless critical, but we'll keep the counter.
	// actually, let's just use fmt for the New call since it is before logger init usually
	//fmt.Printf("DEBUG: DevBrowser New Instance #%d created\n", id)

	return browser
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

	h.OpenBrowser(h.lastPort, h.lastHttps)

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

func (b *DevBrowser) SetTestMode(testMode bool) {
	b.testMode = testMode
}

// monitorBrowserClose monitors the browser context and updates state when browser is closed manually
func (b *DevBrowser) monitorBrowserClose() {
	b.mu.Lock()
	ctx := b.ctx
	b.mu.Unlock()

	if ctx == nil {
		return
	}

	// Wait for context to be done (browser closed)
	<-ctx.Done()

	b.mu.Lock()
	defer b.mu.Unlock()

	// Only handle if browser was marked as open (manual close by user)
	if b.isOpen {
		b.Logger("Browser closed by user")
		b.isOpen = false
		b.ctx = nil
		b.cancel = nil
		b.ui.RefreshUI()
	}
}
