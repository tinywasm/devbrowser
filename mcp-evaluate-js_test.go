package devbrowser

import (
	"strings"
	"testing"
)

func TestEvaluateJsPrimitive(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getEvaluateJsTools()[0].Execute(map[string]any{"script": "2 + 2"}, progress)

	output := <-progress
	if output != "4" {
		t.Errorf("Expected '4', got '%s'", output)
	}
}

func TestEvaluateJsString(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getEvaluateJsTools()[0].Execute(map[string]any{"script": "'hello' + ' ' + 'world'"}, progress)

	output := <-progress
	if output != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", output)
	}
}

func TestEvaluateJsObject(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getEvaluateJsTools()[0].Execute(map[string]any{"script": "({a: 1})"}, progress)

	output := <-progress
	if !strings.Contains(output, `"a":1`) {
		t.Errorf("Expected JSON object, got '%s'", output)
	}
}

func TestEvaluateJsError(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getEvaluateJsTools()[0].Execute(map[string]any{"script": "throw new Error('test')"}, progress)

	output := <-progress
	if !strings.HasPrefix(output, "Error:") {
		t.Errorf("Expected error message, got '%s'", output)
	}
}

func TestEvaluateJsBrowserNotOpen(t *testing.T) {
	db, _ := DefaultTestBrowser()

	progress := make(chan string)
	go db.getEvaluateJsTools()[0].Execute(map[string]any{"script": "1+1"}, progress)

	output := <-progress
	if output != "Browser is not open. Please open it first with browser_open" {
		t.Errorf("Expected 'Browser is not open' error, got '%s'", output)
	}
}

func TestEvaluateJsEmptyScript(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	db.OpenBrowser()
	<-db.readyChan

	progress := make(chan string)
	go db.getEvaluateJsTools()[0].Execute(map[string]any{"script": ""}, progress)

	output := <-progress
	if output != "Script parameter is required" {
		t.Errorf("Expected 'Script parameter is required' error, got '%s'", output)
	}
}
