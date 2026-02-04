package devbrowser


import (
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
			"browser_size":     "1280,720",
		},
	}

	exitChan := make(chan bool)
	browser := New(nil, store, exitChan)

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
	browser := New(nil, store, exitChan)

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
	browser := New(nil, store, exitChan)

	// Change values
	browser.position = "500,300"
	browser.width = 1920
	browser.height = 1080

	// Save config
	err := browser.SaveConfig()
	if err != nil {
		t.Fatalf("saveBrowserConfig failed: %v", err)
	}

	// Verify stored values
	if store.data["browser_position"] != "500,300" {
		t.Errorf("Expected stored position '500,300', got '%s'", store.data["browser_position"])
	}

	if store.data["browser_size"] != "1920,1080" {
		t.Errorf("Expected stored size '1920,1080', got '%s'", store.data["browser_size"])
	}
}

// Test that setBrowserPositionAndSize saves to store
func TestSetBrowserPositionAndSizeSavesToStore(t *testing.T) {
	store := &mockStore{
		data: map[string]string{},
	}

	exitChan := make(chan bool)
	browser := New(nil, store, exitChan)

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

	if store.data["browser_size"] != "800,600" {
		t.Errorf("Expected stored size '800,600', got '%s'", store.data["browser_size"])
	}
}

// Test that invalid values don't overwrite valid stored values
func TestLoadBrowserConfigIgnoresInvalidValues(t *testing.T) {
	store := &mockStore{
		data: map[string]string{
			"browser_position": "100,200",
			"browser_size":     "invalid",
		},
	}

	exitChan := make(chan bool)
	browser := New(nil, store, exitChan)

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
	browser1 := New(nil, store, exitChan)
	browser1.position = "1920,0"
	browser1.width = 1600
	browser1.height = 900
	browser1.SaveConfig()

	// Simulate restart - create new browser instance with same store
	exitChan2 := make(chan bool)
	browser2 := New(nil, store, exitChan2)

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
