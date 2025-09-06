package devbrowser

import (
	"errors"
	"fmt"

	"github.com/go-rod/rod"
)

type DevBrowser struct {
	config   serverConfig
	ui       userInterface
	width    int    // ej "800" default "1024"
	height   int    //ej: "600" default "768"
	position string //ej: "1930,0" (when you have second monitor) default: "0,0"

	isOpen bool // Indica si el navegador está abierto

	// Rod fields
	launcherURL string
	browser     *rod.Browser
	page        *rod.Page
	cancelFunc  func() // Custom cancel function

	readyChan chan bool
	errChan   chan error
	exitChan  chan bool

	logger func(message ...any) // For logging output
}

type serverConfig interface {
	GetServerPort() string
}

type userInterface interface {
	ReturnFocus() error
}

/*
devbrowser.New creates a new DevBrowser instance.

	type serverConfig interface {
		GetServerPort() string
	}

	type userInterface interface {
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
	if b.page == nil {
		return errors.New("page not initialized")
	}

	if err := b.page.Navigate(url); err != nil {
		return err
	}
	return nil
}

func (b *DevBrowser) Reload() error {
	if b.page != nil && b.isOpen {
		b.log("Reload")
		if err := b.page.Reload(); err != nil {
			return errors.New("Reload " + err.Error())
		}
	}
	return nil
}

// log writes a simple message to the configured logger (if any).
func (h *DevBrowser) log(msg string) {
	if h == nil || h.logger == nil {
		return
	}
	// add newline for readability
	h.logger(msg + "\n")
}

// logf writes a formatted message to the configured logger (if any).
func (h *DevBrowser) logf(format string, a ...any) {
	if h == nil || h.logger == nil {
		return
	}
	h.logger(fmt.Sprintf(format, a...) + "\n")
}
