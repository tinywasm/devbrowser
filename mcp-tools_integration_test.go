package devbrowser

import (
	"testing"
)

func TestGetMCPToolsMetadata_AllToolsRegistered(t *testing.T) {
	db, _ := DefaultTestBrowser()
	defer db.CloseBrowser()

	tools := db.GetMCPToolsMetadata()

	expectedToolNames := []string{
		"browser_open",
		"browser_close",
		"browser_reload",
		"browser_get_console",
		"browser_screenshot",
		//"browser_evaluate_js",
		//"browser_get_network_logs",
		//"browser_get_errors",
		//"browser_click_element",
	}

	if len(tools) != len(expectedToolNames) {
		t.Errorf("Expected %d tools, but got %d", len(expectedToolNames), len(tools))
	}

	registeredTools := make(map[string]bool)
	for _, tool := range tools {
		registeredTools[tool.Name] = true
	}

	for _, name := range expectedToolNames {
		if !registeredTools[name] {
			t.Errorf("Expected tool '%s' to be registered, but it was not", name)
		}
	}
}
