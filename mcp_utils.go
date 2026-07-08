package devbrowser

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
	"github.com/tinywasm/model"
)

// EmptyInputSchema is the JSON Schema for a tool that takes no arguments.
// MCP clients require inputSchema to be a valid JSON Schema object; an empty
// string or null is rejected and causes the ENTIRE tools/list to be discarded
// (Claude Code validates tools/list with Zod).
const EmptyInputSchema = `{"type":"object","properties":{}}`

// EncodeSchema builds a valid JSON Schema "object" string for an MCP tool's
// inputSchema, derived from the args model's Schema() field metadata.
//
// It replaces the previous broken behavior that json-encoded the struct's zero
// values (e.g. {"fullpage":false}), which is NOT a JSON Schema and is rejected
// by MCP clients. Returns EmptyInputSchema for a nil model or one with no fields.
//
// The parameter is model.Fielder because that is the interface exposing
// Schema() []model.Field; every *XxxArgs type passed by the mcp-*.go files
// satisfies it (they are ormc-generated models). Call sites are unchanged.
func EncodeSchema(m model.Fielder) string {
	if m == nil {
		return EmptyInputSchema
	}
	fields := m.Schema()
	if len(fields) == 0 {
		return EmptyInputSchema
	}
	var b fmt.Conv
	b.Write(`{"type":"object","properties":{`)
	var required []string
	for i, f := range fields {
		if i > 0 {
			b.Write(",")
		}
		b.Write(`"`)
		b.Write(f.Name)
		b.Write(`":`)
		b.Write(jsonSchemaType(f.Type))
		if f.NotNull {
			required = append(required, f.Name)
		}
	}
	b.Write("}")
	if len(required) > 0 {
		b.Write(`,"required":[`)
		for i, name := range required {
			if i > 0 {
				b.Write(",")
			}
			b.Write(`"`)
			b.Write(name)
			b.Write(`"`)
		}
		b.Write("]")
	}
	b.Write("}")
	return b.String()
}

// EncodeArgs is a helper for testing that encodes actual argument values to JSON.
func EncodeArgs(f model.Encodable) string {
	var s string
	_ = json.Encode(f, &s)
	return s
}

// jsonSchemaType maps a model.FieldType to its JSON Schema fragment.
// Deterministic mapping (see model.FieldType docs):
//   FieldText, FieldRaw, FieldBlob -> string
//   FieldInt                       -> integer
//   FieldFloat                     -> number
//   FieldBool                      -> boolean
//   FieldIntSlice                  -> array of integer
//   FieldStruct                    -> object
//   FieldStructSlice               -> array of object
func jsonSchemaType(t model.FieldType) string {
	switch t {
	case model.FieldInt:
		return `{"type":"integer"}`
	case model.FieldFloat:
		return `{"type":"number"}`
	case model.FieldBool:
		return `{"type":"boolean"}`
	case model.FieldIntSlice:
		return `{"type":"array","items":{"type":"integer"}}`
	case model.FieldStruct:
		return `{"type":"object"}`
	case model.FieldStructSlice:
		return `{"type":"array","items":{"type":"object"}}`
	default: // FieldText, FieldRaw, FieldBlob
		return `{"type":"string"}`
	}
}
