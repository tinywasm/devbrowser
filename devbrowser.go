package devbrowser

import (
	"context"
	"errors"

	"github.com/chromedp/chromedp"
)

type DevBrowser struct {
	config   serverConfig
	ui       userInterface
	width    int    // ej "800" default "1024"
	height   int    //ej: "600" default "768"
	position string //ej: "1930,0" (when you have second monitor) default: "0,0"
	headless bool   // true para modo headless (sin UI), false muestra el navegador

	isOpen bool // Indica si el navegador está abierto

	// chromedp fields
	ctx    context.Context
	cancel context.CancelFunc

	readyChan chan bool
	errChan   chan error
	exitChan  chan bool

	logger func(message ...any) // For logging output

	lastOpID string // For tracking last operation ID
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
func New(sc serverConfig, ui userInterface, exitChan chan bool, logger func(message ...any)) *DevBrowser {

	browser := &DevBrowser{
		config:    sc,
		ui:        ui,
		width:     1024,  // Default width
		height:    768,   // Default height
		position:  "0,0", // Default position
		readyChan: make(chan bool),
		errChan:   make(chan error),
		exitChan:  exitChan,
		logger:    logger,
	}
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
		b.logger("Reload")
		if err := chromedp.Run(b.ctx, chromedp.Reload()); err != nil {
			return errors.New("Reload " + err.Error())
		}
	}
	return nil
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
		b.logger("Browser closed by user")
		b.isOpen = false
		b.ctx = nil
		b.cancel = nil
		b.ui.RefreshUI()
	}
}
