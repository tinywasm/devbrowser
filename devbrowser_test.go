package devbrowser

import "testing"

// minimal stub implementations for interfaces used by New
type fakeServerConfig struct{}

func (f fakeServerConfig) GetServerPort() string { return "8080" }

type fakeUI struct{}

func (f fakeUI) ReturnFocus() error { return nil }

func TestNewDefaults(t *testing.T) {
	db := New(fakeServerConfig{}, fakeUI{}, make(chan bool), nil)
	if db == nil {
		t.Fatal("New returned nil")
	}
	if db.width != 1024 {
		t.Fatalf("expected default width 1024, got %d", db.width)
	}
	if db.height != 768 {
		t.Fatalf("expected default height 768, got %d", db.height)
	}
	if db.position != "0,0" {
		t.Fatalf("expected default position '0,0', got '%s'", db.position)
	}
}

func TestCloseBrowserWhenClosed(t *testing.T) {
	db := New(fakeServerConfig{}, fakeUI{}, make(chan bool), nil)
	// Ensure isOpen is false by default
	if db.isOpen {
		t.Fatal("expected isOpen false by default")
	}
	if err := db.CloseBrowser(); err == nil {
		t.Fatal("expected CloseBrowser to return error when already closed")
	}
}
