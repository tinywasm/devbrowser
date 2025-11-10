package devbrowser

import (
	"strconv"
	"testing"
)

// mockStore for testing
type mockStore struct {
	data map[string]string
}

func (m *mockStore) Get(key string) (string, error) {
	val, ok := m.data[key]
	if !ok {
		return "", nil
	}
	return val, nil
}

func (m *mockStore) Set(key, value string) error {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
	return nil
}

// Test that browser config is loaded from store on New()
func TestLoadBrowserConfigOnNew(t *testing.T) {
	store := &mockStore{
		data: map[string]string{
			"browser_position": "100,200",
			"browser_width":    "1280",
			"browser_height":   "720",
		},
	}

	exitChan := make(chan bool)
	browser := New(nil, nil, store, exitChan, nil)

	if browser.position != "100,200" {
		t.Errorf("Expected position '100,200', got '%s'", browser.position)
	}

	if browser.width != 1280 {
		t.Errorf("Expected width 1280, got %d", browser.width)
	}

	if browser.height != 720 {
		t.Errorf("Expected height 720, got %d", browser.height)
	}
}

// Test that browser config uses defaults when store is empty
func TestLoadBrowserConfigDefaults(t *testing.T) {
	store := &mockStore{
		data: map[string]string{},
	}

	exitChan := make(chan bool)
	browser := New(nil, nil, store, exitChan, nil)

	if browser.position != "0,0" {
		t.Errorf("Expected default position '0,0', got '%s'", browser.position)
	}

	if browser.width != 1024 {
		t.Errorf("Expected default width 1024, got %d", browser.width)
	}

	if browser.height != 768 {
		t.Errorf("Expected default height 768, got %d", browser.height)
	}
}

// Test that saveBrowserConfig stores values correctly
func TestSaveBrowserConfig(t *testing.T) {
	store := &mockStore{
		data: map[string]string{},
	}

	exitChan := make(chan bool)
	browser := New(nil, nil, store, exitChan, nil)

	// Change values
	browser.position = "500,300"
	browser.width = 1920
	browser.height = 1080

	// Save config
	err := browser.saveBrowserConfig()
	if err != nil {
		t.Fatalf("saveBrowserConfig failed: %v", err)
	}

	// Verify stored values
	if store.data["browser_position"] != "500,300" {
		t.Errorf("Expected stored position '500,300', got '%s'", store.data["browser_position"])
	}

	if store.data["browser_width"] != "1920" {
		t.Errorf("Expected stored width '1920', got '%s'", store.data["browser_width"])
	}

	if store.data["browser_height"] != "1080" {
		t.Errorf("Expected stored height '1080', got '%s'", store.data["browser_height"])
	}
}

// Test that setBrowserPositionAndSize saves to store
func TestSetBrowserPositionAndSizeSavesToStore(t *testing.T) {
	store := &mockStore{
		data: map[string]string{},
	}

	exitChan := make(chan bool)
	browser := New(nil, nil, store, exitChan, nil)

	// Set new position and size
	err := browser.setBrowserPositionAndSize("1930,0:800,600")
	if err != nil {
		t.Fatalf("setBrowserPositionAndSize failed: %v", err)
	}

	// Verify browser values
	if browser.position != "1930,0" {
		t.Errorf("Expected position '1930,0', got '%s'", browser.position)
	}

	if browser.width != 800 {
		t.Errorf("Expected width 800, got %d", browser.width)
	}

	if browser.height != 600 {
		t.Errorf("Expected height 600, got %d", browser.height)
	}

	// Verify stored values
	if store.data["browser_position"] != "1930,0" {
		t.Errorf("Expected stored position '1930,0', got '%s'", store.data["browser_position"])
	}

	width, _ := strconv.Atoi(store.data["browser_width"])
	if width != 800 {
		t.Errorf("Expected stored width 800, got %d", width)
	}

	height, _ := strconv.Atoi(store.data["browser_height"])
	if height != 600 {
		t.Errorf("Expected stored height 600, got %d", height)
	}
}

// Test that invalid values don't overwrite valid stored values
func TestLoadBrowserConfigIgnoresInvalidValues(t *testing.T) {
	store := &mockStore{
		data: map[string]string{
			"browser_position": "100,200",
			"browser_width":    "invalid",
			"browser_height":   "not_a_number",
		},
	}

	exitChan := make(chan bool)
	browser := New(nil, nil, store, exitChan, nil)

	// Position should load correctly
	if browser.position != "100,200" {
		t.Errorf("Expected position '100,200', got '%s'", browser.position)
	}

	// Width and height should use defaults due to invalid values
	if browser.width != 1024 {
		t.Errorf("Expected default width 1024, got %d", browser.width)
	}

	if browser.height != 768 {
		t.Errorf("Expected default height 768, got %d", browser.height)
	}
}

// Test persistence across browser restarts
func TestBrowserConfigPersistenceAcrossRestarts(t *testing.T) {
	store := &mockStore{
		data: map[string]string{},
	}

	exitChan := make(chan bool)

	// First browser instance
	browser1 := New(nil, nil, store, exitChan, nil)
	browser1.position = "1920,0"
	browser1.width = 1600
	browser1.height = 900
	browser1.saveBrowserConfig()

	// Simulate restart - create new browser instance with same store
	exitChan2 := make(chan bool)
	browser2 := New(nil, nil, store, exitChan2, nil)

	// Verify values were loaded from store
	if browser2.position != "1920,0" {
		t.Errorf("Expected position '1920,0' after restart, got '%s'", browser2.position)
	}

	if browser2.width != 1600 {
		t.Errorf("Expected width 1600 after restart, got %d", browser2.width)
	}

	if browser2.height != 900 {
		t.Errorf("Expected height 900 after restart, got %d", browser2.height)
	}
}
