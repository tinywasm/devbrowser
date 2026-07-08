# PLAN — Generar `inputSchema` JSON Schema válido en las MCP tools de devbrowser

> This plan is dispatched via the CodeJob workflow. See skill: `agents-workflow`.

You are an external agent with **zero prior context** about this project. Everything you
need is in this file. Read it fully before writing code.

---

## 1. Problema

Todas las MCP tools de devbrowser (`browser_screenshot`, `browser_click_element`,
`browser_get_console`, … 17 en total) exponen un `inputSchema` **inválido**. Un cliente MCP
como Claude Code valida la respuesta de `tools/list` contra un esquema JSON Schema (Zod). Si
**una sola** tool tiene `inputSchema` inválido, el cliente **descarta el array COMPLETO** de
tools → el servidor MCP aparece "Connected" pero el agente **no ve ninguna tool**.

Evidencia real (log de Claude Code):

```
Failed to fetch tools: [
  { "path": ["tools", 3, "inputSchema", "type"], "message": "Invalid input: expected \"object\"" },
  { "path": ["tools", 18, "inputSchema"], "message": "Invalid input: expected object, received null" },
  ...
]
```

### Causa raíz

El helper `EncodeSchema` en `mcp_utils.go` **serializa el struct de argumentos con sus valores
cero** en vez de generar un JSON Schema:

```go
// mcp_utils.go  (ROTO)
func EncodeSchema(f model.Encodable) string {
	var s string
	_ = json.Encode(f, &s)   // ❌ produce {"fullpage":false}, no un JSON Schema
	return s
}
```

Para `browser_screenshot` esto emite `{"fullpage":false}`. Un `inputSchema` **válido** debe ser
un objeto JSON Schema:

```json
{"type":"object","properties":{"fullpage":{"type":"boolean"}}}
```

Todos los tipos de argumentos (`ScreenshotArgs`, `ClickElementArgs`, …) son modelos generados
por ormc y **ya exponen su metadata de campos** mediante el método
`Schema() []model.Field`, p.ej.:

```go
var _schemaScreenshotArgs = []model.Field{
	{Name: "fullpage", Type: model.FieldBool},
}
func (m *ScreenshotArgs) Schema() []model.Field { return _schemaScreenshotArgs }
```

Por lo tanto el JSON Schema válido se **genera a partir de `Schema()`**, sin reflection y sin
hardcodear nada.

---

## 2. Objetivo

Reescribir **una sola función** (`EncodeSchema` en `mcp_utils.go`) para que construya un JSON
Schema `object` válido a partir de `Schema() []model.Field`. Con eso las 17 tools quedan
corregidas de golpe — **no hay que tocar los 17 archivos `mcp-*.go`** (todos llaman a
`EncodeSchema`, que se mantiene con el mismo nombre y sitios de llamada).

---

## 3. Cambios

### 3.1 `mcp_utils.go` — reescribir `EncodeSchema` + añadir mapeo de tipos

**Reglas del ecosistema TinyWasm (OBLIGATORIAS):**
- **Sin stdlib** en paquetes compilables a WASM: usar `github.com/tinywasm/fmt` (NO `strings`,
  `strconv`, `bytes`). El buffer es `fmt.Conv` con métodos `.Write(string)` y `.String()`
  (ya se usa así en el ecosistema).
- **Sin strings mágicos duplicados**: los fragmentos de schema van en la función `jsonSchemaType`
  como único punto de verdad; la constante del schema vacío es una constante nombrada.

Reemplazar el contenido de `mcp_utils.go` por:

```go
package devbrowser

import (
	"github.com/tinywasm/fmt"
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
```

> NOTA sobre la firma: si algún call site pasa un tipo que satisface `model.Encodable` pero no
> `model.Fielder`, ajusta ese call site para pasar el `*XxxArgs` concreto (todos son modelos ormc
> con `Schema()`). NO reintroduzcas la serialización del struct.

### 3.2 Verificar los 17 call sites

No requieren cambios de lógica (siguen llamando `EncodeSchema(new(XxxArgs))`). Solo confirma que
el proyecto compila; si el compilador se queja por el cambio de tipo del parámetro
(`model.Encodable` → `model.Fielder`), corrige el call site pasando el mismo `new(XxxArgs)` (ya
lo satisface) — normalmente no hace falta cambiar nada.

---

## 4. Tests

Añade `mcp_utils_test.go` (paquete `devbrowser`) que verifique:

1. `EncodeSchema(new(ScreenshotArgs))` produce exactamente:
   `{"type":"object","properties":{"fullpage":{"type":"boolean"}}}`
2. `EncodeSchema(nil)` devuelve `EmptyInputSchema`.
3. Para **cada** tool devuelta por `GetMCPTools()`, su `InputSchema`:
   - No es `""` ni `"null"`.
   - Decodifica como JSON válido (usa `github.com/tinywasm/json` `Decode`).
   - Contiene la subcadena `"type":"object"` en la raíz.

Ejecutar: `go test ./...` (o `gotest ./...` si está disponible). Todos deben pasar.

---

## 5. Documentación

- Si existe `docs/ARCHITECTURE.md` o `README.md` con una sección de MCP tools que describa cómo
  se genera el `inputSchema`, actualízala: ahora se deriva de `Schema()` como JSON Schema válido.
- Si no existe tal sección, no crees documentación nueva.

---

## Reglas de calidad (recordatorio)

- Sin stdlib en código WASM: `tinywasm/fmt`, `tinywasm/json`, `tinywasm/model` únicamente.
- Sin literales string repetidos en la lógica: fragmentos de schema centralizados en
  `jsonSchemaType`; schema vacío en la constante `EmptyInputSchema`.
- No introducir `encoding/json`, `reflect`, `strings`, `strconv`, `bytes`.

---

## Stages

| # | Stage | Output |
|---|-------|--------|
| 1 | Reescribir `EncodeSchema` + `jsonSchemaType` + `EmptyInputSchema` en `mcp_utils.go` | JSON Schema válido desde `Schema()` |
| 2 | Compilar y ajustar call sites si el tipo del parámetro lo requiere | build verde |
| 3 | Añadir `mcp_utils_test.go` con las aserciones de §4 | tests verdes |
| 4 | Actualizar docs de MCP tools si existen | docs consistentes |
