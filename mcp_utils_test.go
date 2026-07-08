package devbrowser

import (
	"testing"
	"github.com/tinywasm/fmt"
)

func TestEncodeSchema(t *testing.T) {
	// 1. EncodeSchema(new(ScreenshotArgs)) produces exactly the expected JSON Schema.
	got := EncodeSchema(new(ScreenshotArgs))
	want := `{"type":"object","properties":{"fullpage":{"type":"boolean"}}}`
	if got != want {
		t.Errorf("EncodeSchema(new(ScreenshotArgs)) = %v, want %v", got, want)
	}

	// 2. EncodeSchema(nil) returns EmptyInputSchema.
	if got := EncodeSchema(nil); got != EmptyInputSchema {
		t.Errorf("EncodeSchema(nil) = %v, want %v", got, EmptyInputSchema)
	}

	// 3. For each tool returned by GetMCPTools(), its InputSchema contains "type":"object".
	b := &DevBrowser{}
	tools := b.GetMCPTools()
	if len(tools) == 0 {
		t.Fatal("No tools found")
	}

	for _, tool := range tools {
		if tool.InputSchema == "" || tool.InputSchema == "null" {
			t.Errorf("Tool %s has invalid InputSchema: %q", tool.Name, tool.InputSchema)
			continue
		}

		if !fmt.Contains(tool.InputSchema, `{"type":"object"`) {
			t.Errorf("Tool %s InputSchema root type is likely not object: %v", tool.Name, tool.InputSchema)
		}
	}
}
