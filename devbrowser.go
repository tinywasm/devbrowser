package devbrowser

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"golang.design/x/clipboard"
)

type store interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type DevBrowser struct {
	config   serverConfig
	ui       userInterface
	width    int    // ej "800" default "1024"
	height   int    //ej: "600" default "768"
	position string //ej: "1930,0" (when you have second monitor) default: "0,0"
	headless bool   // true para modo headless (sin UI), false muestra el navegador

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

	// Load stored position and size from db
	browser.loadBrowserConfig()

	return browser
}

// loadBrowserConfig loads position and size from the store
func (b *DevBrowser) loadBrowserConfig() {
	// Load position
	if pos, err := b.db.Get("browser_position"); err == nil && pos != "" {
		b.position = pos
	}

	// Load width
	if w, err := b.db.Get("browser_width"); err == nil && w != "" {
		if width, err := strconv.Atoi(w); err == nil {
			b.width = width
		}
	}

	// Load height
	if h, err := b.db.Get("browser_height"); err == nil && h != "" {
		if height, err := strconv.Atoi(h); err == nil {
			b.height = height
		}
	}
}

// saveBrowserConfig saves current position and size to the store
func (b *DevBrowser) saveBrowserConfig() error {
	// Save position
	if err := b.db.Set("browser_position", b.position); err != nil {
		return err
	}

	// Save width
	if err := b.db.Set("browser_width", strconv.Itoa(b.width)); err != nil {
		return err
	}

	// Save height
	if err := b.db.Set("browser_height", strconv.Itoa(b.height)); err != nil {
		return err
	}

	return nil
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
