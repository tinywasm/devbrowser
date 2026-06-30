# PLAN — Nuevas herramientas MCP para devbrowser

## Prerequisito — instalar gotest

Los agentes externos corren en entornos aislados donde `gotest` no está disponible globalmente. Ejecutar **antes de cualquier otra tarea**:

```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

Verificar instalación:
```bash
gotest -h
```

Correr la suite completa (desde la raíz del módulo `devbrowser`):
```bash
cd /home/cesar/Dev/Project/tinywasm/devbrowser && gotest
```

---

## Contexto de implementación

### Paquete
```
package devbrowser
```
Módulo Go: `github.com/tinywasm/devbrowser` (ver `go.mod`).

### Patrón obligatorio para cada tool

**1. Args struct** — agregar en `models.go`:
```go
type NombreArgs struct {
    Campo string `input:"-"`       // string libre
    Num   int    `input:"number"`  // numérico
    Flag  bool   `input:"-"`       // booleano
}
```

**2. Archivo de implementación** — un archivo `mcp-<nombre>.go` por grupo:
```go
package devbrowser

import (
    "fmt"
    "github.com/tinywasm/context"
    "github.com/tinywasm/devbrowser/chromedp"
    "github.com/tinywasm/mcp"
)

func (b *DevBrowser) GetNombreTools() []mcp.Tool {
    return []mcp.Tool{
        {
            Name:        "browser_nombre",
            Description: "...",
            InputSchema: EncodeSchema(new(NombreArgs)),
            Resource:    "browser",
            Action:      'r',  // 'r'=solo lectura, 'u'=modifica estado
            Execute: func(ctx *context.Context, req mcp.Request) (*mcp.Result, error) {
                if !b.IsOpenFlag {
                    return nil, fmt.Errorf("Browser is not open. Please open it first with browser_open")
                }
                var args NombreArgs
                if err := req.Bind(&args); err != nil {
                    return nil, err
                }
                // ... lógica
                return mcp.Text(resultado), nil
            },
        },
    }
}
```

**3. Registro en `mcp-tools.go`** — agregar la línea en `GetMCPTools()`:
```go
tools = append(tools, b.GetNombreTools()...)
```

### Ejecutar JS y obtener string
```go
var result string
err := chromedp.Run(b.Ctx,
    chromedp.Evaluate(`/* js aquí */`, &result),
)
```

### Ejecutar JS que devuelve Promise (fetch, etc.)
```go
import "github.com/tinywasm/devbrowser/cdproto/runtime"

var result interface{}
err := chromedp.Run(b.Ctx,
    chromedp.Evaluate(jsCode, &result, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
        return p.WithAwaitPromise(true)
    }),
)
```

---

## Herramientas implementadas (no tocar)

`browser_navigate`, `browser_get_content`, `browser_screenshot`, `browser_inspect_element`,
`browser_evaluate_js`, `browser_get_network_logs`, `browser_get_console`, `browser_get_errors`,
`browser_get_performance`, `browser_fill_element`, `browser_click_element`, `browser_swipe_element`,
`browser_emulate_device`

---

## Herramientas a implementar

---

### TASK-1 — `browser_get_source`

**Archivo:** `mcp-source.go`
**Args en `models.go`:**
```go
type GetSourceArgs struct {
    Selector string `input:"-"` // vacío = página completa
}
```

**Registro en `mcp-tools.go`:**
```go
tools = append(tools, b.GetSourceTools()...)
```

**Lógica:**
```go
var js string
if args.Selector == "" {
    js = `document.documentElement.outerHTML`
} else {
    js = fmt.Sprintf(`document.querySelector(%q)?.outerHTML ?? "Element not found"`, args.Selector)
}
var result string
err := chromedp.Run(b.Ctx, chromedp.Evaluate(js, &result))
```

**Action:** `'r'`

**Por qué:** `browser_get_content` devuelve HTML semántico simplificado generado por JS — pierde el `<head>`, atributos `style=""` inline, y la estructura real del DOM. Esta tool devuelve el HTML crudo tal como está en el DOM, necesario para ingeniería inversa fiel.

---

### TASK-2 — `browser_get_styles`

**Archivo:** `mcp-styles.go`
**Args en `models.go`:**
```go
type GetStylesArgs struct {
    Selector string `input:"-"` // vacío = todas las reglas de todos los stylesheets
    Sheet    int    `input:"number"` // índice de stylesheet (-1 = todos)
}
```

**Registro en `mcp-tools.go`:**
```go
tools = append(tools, b.GetStylesTools()...)
```

**Lógica JS (evaluar como string):**
```javascript
// Todas las reglas de todos los stylesheets
[...document.styleSheets].flatMap(s => {
  try { return [...s.cssRules].map(r => r.cssText) }
  catch(e) { return ['/* cross-origin: ' + s.href + ' */'] }
}).join('\n')
```

Si `Selector != ""`, agregar al final del JS filtrado por `selectorText`:
```javascript
[...document.styleSheets].flatMap(s => {
  try { return [...s.cssRules] }
  catch(e) { return [] }
}).filter(r => r.selectorText && r.selectorText.includes(SELECTOR))
  .map(r => r.cssText).join('\n')
```

**Action:** `'r'`

**Por qué:** No hay forma de extraer las reglas CSS originales de los archivos `.css` cargados. `browser_inspect_element` solo da estilos computados de un elemento instanciado. Esta tool expone las reglas fuente, necesarias para replicar el diseño.

---

### TASK-3 — `browser_get_storage`

**Archivo:** `mcp-storage.go`
**Args en `models.go`:**
```go
type GetStorageArgs struct {
    Type string `input:"-"` // "local" | "session" | "cookies" (default: "local")
}
```

**Registro en `mcp-tools.go`:**
```go
tools = append(tools, b.GetStorageTools()...)
```

**Lógica por tipo:**
```go
var js string
switch args.Type {
case "session":
    js = `JSON.stringify(Object.fromEntries(Object.entries(sessionStorage)))`
case "cookies":
    js = `document.cookie`
default: // "local"
    js = `JSON.stringify(Object.fromEntries(Object.entries(localStorage)))`
}
var result string
err := chromedp.Run(b.Ctx, chromedp.Evaluate(js, &result))
```

**Action:** `'r'`

**Por qué:** Los sistemas web persisten tokens de sesión, configuración de usuario, y estado en storage del browser. Sin acceder a estos datos no se puede entender ni replicar el flujo de autenticación.

---

### TASK-4 — `browser_get_asset`

**Archivo:** `mcp-asset.go`
**Args en `models.go`:**
```go
type GetAssetArgs struct {
    URL string `input:"-"` // URL absoluta del archivo JS o CSS a descargar
}
```

**Registro en `mcp-tools.go`:**
```go
tools = append(tools, b.GetAssetTools()...)
```

**Lógica** (fetch desde el contexto del browser — hereda cookies y sesión activa):
```go
import "github.com/tinywasm/devbrowser/cdproto/runtime"

js := fmt.Sprintf(`fetch(%q).then(r => r.text())`, args.URL)
var result interface{}
err := chromedp.Run(b.Ctx,
    chromedp.Evaluate(js, &result, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
        return p.WithAwaitPromise(true)
    }),
)
// castear result a string
```

**Action:** `'r'`

**Por qué:** `browser_get_network_logs` muestra URLs de archivos cargados pero no su contenido. Hacer `browser_evaluate_js` con un `fetch()` manual es tedioso y tiene límites. Esta tool simplifica la descarga del fuente de archivos JS/CSS usando la sesión ya activa del browser (evita problemas de CORS y autenticación).

---

### TASK-5 — `browser_intercept_request`

**Archivo:** `mcp-intercept.go`
**Args en `models.go`:**
```go
type InterceptRequestArgs struct {
    Action string `input:"-"` // "start" | "stop" | "get"
    Filter string `input:"-"` // filtro de URL (substring), vacío = todo
    Limit  int    `input:"number"`
}
```

**Registro en `mcp-tools.go`:**
```go
tools = append(tools, b.GetInterceptTools()...)
```

**Lógica:**
Requiere CDP `Fetch` domain. Agregar en `DevBrowser` struct (en `devbrowser.go` o `models_orm.go`):
```go
InterceptActive  bool
InterceptedReqs  []InterceptedRequest  // definir struct con URL, Method, RequestBody, ResponseBody, Status
InterceptMutex   sync.Mutex
```

Nuevo struct en `models.go`:
```go
type InterceptedRequest struct {
    URL          string
    Method       string
    RequestBody  string
    ResponseBody string
    Status       int
}
```

Activar con `chromedp.ListenTarget` + `fetch.Enable()` + `fetch.ContinueRequest()` para no bloquear el tráfico. Capturar `fetch.EventRequestPaused` para leer body de request y response.

**Action:** `'u'` para start/stop, `'r'` para get — implementar como acción única con switch en `args.Action`.

**Por qué:** La carencia más crítica. `browser_get_network_logs` solo captura metadata (URL, status, timing). Para ingeniería inversa de APIs internas se necesitan los **bodies** de requests y responses — qué parámetros recibe cada endpoint y qué JSON devuelve. Sin esto hay que deducir la API solo observando el comportamiento visual.

**Nota de complejidad:** Es la tool más compleja. Requiere manejar un loop de eventos CDP en goroutine separada y gestionar el ciclo de vida del interceptor. Implementar después de las TASK-1 a TASK-4.

---

## Tests

### Convenciones obligatorias

- Paquete: `package devbrowser_test`
- Ubicación: `tests/<nombre>_test.go`
- Sin librerías externas — solo `testing`, `net/http/httptest`, `strings`
- Usar siempre `DefaultTestBrowser()` definido en `tests/default_test.go`
- Servidor local con `httptest.NewServer` para aislar el browser del exterior
- No hacer matching exacto de strings HTML completos — verificar substrings significativos

### TASK-1 test — `tests/mcp_source_test.go`

Verificar que `browser_get_source` devuelve HTML crudo con elementos que `browser_get_content` omite.

```go
func TestGetSource_FullPage(t *testing.T) {
    // HTML con inline style y <head> con <link>
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, `<!DOCTYPE html><html><head><link rel="stylesheet" href="/app.css"></head>
        <body><div style="color:red" class="box">hello</div></body></html>`)
    }))
    defer ts.Close()

    db, _ := DefaultTestBrowser()
    defer db.CloseBrowser()
    if err := db.OpenBrowser(ts.URL, false); err != nil {
        t.Fatal(err)
    }

    result := // llamar GetSourceTools Execute con Selector=""
    // Verificar que contiene el <head>, el inline style y la clase
    if !strings.Contains(result, `<link rel="stylesheet"`) { t.Error("missing <head> content") }
    if !strings.Contains(result, `style="color:red"`) { t.Error("missing inline style") }
    if !strings.Contains(result, `class="box"`) { t.Error("missing class") }
}

func TestGetSource_Selector(t *testing.T) {
    // Verificar que con selector solo devuelve el fragmento del elemento
    // result NO debe contener <html> ni <head>
    if strings.Contains(result, "<html") { t.Error("should return only the element, not full page") }
}
```

### TASK-2 test — `tests/mcp_styles_test.go`

Verificar que `browser_get_styles` extrae reglas CSS de `<style>` embebido.

```go
func TestGetStyles_AllSheets(t *testing.T) {
    // Página con <style>.box { color: red; font-size: 14px; }</style>
    // Verificar que result contiene ".box" y "color: red"
}

func TestGetStyles_Selector(t *testing.T) {
    // Con Selector=".box" solo deben aparecer reglas que incluyan ".box"
    // Reglas de otros selectores NO deben aparecer
}

func TestGetStyles_CrossOrigin(t *testing.T) {
    // Stylesheet de origen externo → debe aparecer comentario "/* cross-origin: ... */"
    // en lugar de error fatal
}
```

### TASK-3 test — `tests/mcp_storage_test.go`

Verificar lectura de localStorage, sessionStorage y cookies.

```go
func TestGetStorage_Local(t *testing.T) {
    // Página con: localStorage.setItem("token", "abc123")
    // Verificar que result contiene "token" y "abc123"
}

func TestGetStorage_Session(t *testing.T) {
    // sessionStorage.setItem("step", "2")
    // Verificar presencia de "step" y "2"
}

func TestGetStorage_Cookies(t *testing.T) {
    // document.cookie = "user=cesar"
    // Verificar que result contiene "user=cesar"
}

func TestGetStorage_Empty(t *testing.T) {
    // Storage vacío → result debe ser "{}" o "" sin error
}
```

### TASK-4 test — `tests/mcp_asset_test.go`

Verificar descarga de archivo JS/CSS usando `fetch` desde el contexto del browser.

```go
func TestGetAsset_JS(t *testing.T) {
    // Servidor con /app.js que devuelve "function hello(){}"
    // Navegar a la página, llamar GetAssetTools con URL /app.js
    // Verificar que result contiene "function hello"
}

func TestGetAsset_CSS(t *testing.T) {
    // Servidor con /style.css que devuelve ".btn { padding: 8px; }"
    // Verificar que result contiene ".btn"
}

func TestGetAsset_NotFound(t *testing.T) {
    // URL que devuelve 404 → debe retornar error o string con indicación de fallo
    // No debe panic ni colgar
}
```

### TASK-5 test — `tests/mcp_intercept_test.go`

Verificar captura de body de request y response en llamadas XHR/fetch.

```go
func TestInterceptRequest_CapturesBody(t *testing.T) {
    // Servidor con POST /api que lee body y devuelve JSON
    // Desde la página: fetch('/api', {method:'POST', body:'{"x":1}'})
    // Activar intercept → navegar → ejecutar fetch via evaluate_js → get
    // Verificar que InterceptedReqs contiene RequestBody con "x":1
    //   y ResponseBody con el JSON de respuesta
}

func TestInterceptRequest_StartStop(t *testing.T) {
    // Start → hacer request → Stop → hacer otro request
    // Solo el primer request debe aparecer en los interceptados
}

func TestInterceptRequest_Filter(t *testing.T) {
    // Dos endpoints /api/a y /api/b → Filter="/api/a"
    // Solo debe aparecer el request a /api/a
}
```

### Actualizar `tests/mcp-tools_integration_test.go`

Agregar los 5 nombres nuevos al slice `expectedToolNames`:

```go
"browser_get_source",
"browser_get_styles",
"browser_get_storage",
"browser_get_asset",
"browser_intercept_request",
```

Y actualizar el conteo esperado de `len(tools)`.

### Correr tests

```bash
cd /home/cesar/Dev/Project/tinywasm/devbrowser && gotest
# O test específico:
gotest -run TestGetSource
gotest -run TestGetStyles
gotest -run TestGetStorage
gotest -run TestGetAsset
gotest -run TestInterceptRequest
```

---

## Documentación

### Archivos a actualizar

| Archivo | Qué agregar |
|---|---|
| `README.md` | Las 5 tools nuevas en la tabla de herramientas MCP disponibles |
| `docs/PLAN.md` | Marcar cada TASK como completada al terminarla (agregar `✓` al título) |

### Formato de entrada en README.md

Localizar la tabla existente de tools MCP y agregar una fila por cada tool nueva. Seguir el formato de las filas ya existentes. Ejemplo:

```markdown
| `browser_get_source`         | Obtiene el HTML crudo (outerHTML) de la página completa o de un elemento por selector |
| `browser_get_styles`         | Extrae reglas CSS de los stylesheets cargados, con filtro por selector opcional       |
| `browser_get_storage`        | Lee localStorage, sessionStorage o cookies del dominio actual                         |
| `browser_get_asset`          | Descarga el contenido de un archivo JS o CSS por URL usando la sesión activa          |
| `browser_intercept_request`  | Captura bodies de requests y responses XHR/fetch en vuelo (CDP Fetch domain)         |
```

### Al terminar cada TASK

Marcar en este PLAN el título de la tarea con `✓`:
```
### TASK-1 ✓ — browser_get_source
```
