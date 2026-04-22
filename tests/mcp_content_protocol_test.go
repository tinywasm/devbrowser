package devbrowser_test

import (
	"strings"
	"testing"

	"github.com/tinywasm/context"
	"github.com/tinywasm/mcp"
)

// helpers

func findTool(tools []mcp.Tool, name string) *mcp.Tool {
	for i := range tools {
		if tools[i].Name == name {
			return &tools[i]
		}
	}
	return nil
}

func emptyReq(name string) mcp.Request {
	return mcp.Request{
		Params: mcp.CallToolParams{Name: name, Arguments: "{}"},
		Action: 'r',
	}
}

func contentIsArray(result *mcp.Result) bool {
	return strings.HasPrefix(strings.TrimSpace(result.Content), "[")
}

// Bug C: mcp.Text() devuelve Content como objeto en lugar de array.
// Estos tests fallan con mcp < fix y pasan automáticamente al actualizar la dependencia.

// TestMCPContent_GetErrors_Empty_IsArray verifica que browser_get_errors devuelva content como array
// incluso en el caso vacío "No JavaScript errors captured" — sin necesitar browser con contexto real.
// Este test cubre el path más simple de mcp.Text() en devbrowser.
func TestMCPContent_GetErrors_Empty_IsArray(t *testing.T) {
	db, _ := DefaultTestBrowser()
	db.IsOpenFlag = true
	tool := findTool(db.GetErrorTools(), "browser_get_errors")
	if tool == nil {
		t.Fatal("browser_get_errors tool not found")
	}

	var ctx context.Context
	result, err := tool.Execute(&ctx, emptyReq("browser_get_errors"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !contentIsArray(result) {
		t.Fatalf("Bug C: content is not a JSON array — mcp.Text() fix not applied.\nContent: %s", result.Content)
	}
}

// TestMCPContent_GetNetworkLogs_IsArray verifica que browser_get_network_logs devuelva content como array.
// Con IsOpenFlag=true y sin logs reales, la tool devuelve un mensaje vacío via mcp.Text().
func TestMCPContent_GetNetworkLogs_IsArray(t *testing.T) {
	db, _ := DefaultTestBrowser()
	db.IsOpenFlag = true
	tool := findTool(db.GetNetworkTools(), "browser_get_network_logs")
	if tool == nil {
		t.Fatal("browser_get_network_logs tool not found")
	}

	var ctx context.Context
	result, err := tool.Execute(&ctx, emptyReq("browser_get_network_logs"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !contentIsArray(result) {
		t.Fatalf("Bug C: content is not a JSON array — mcp.Text() fix not applied.\nContent: %s", result.Content)
	}
}
