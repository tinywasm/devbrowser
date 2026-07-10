# PLAN — devbrowser: whitelists explícitas en los campos interpretados por el browser (tests de interacción rojos) + tui.go → HandlerSelection

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Parte de `tinywasm/docs/MCP_DAEMON_HARDENING_MASTER_PLAN.md`.
> **Gate etapa 1: CERRADO** — `tinywasm/model` v0.0.8 publicado (2026-07-10):
> la whitelist positiva del Field reemplaza el piso del kind `Text()` base
> (contrato en `model/docs/ARCHITECTURE.md` §8). Sin gates pendientes.
> Idioma: español (decisión del mantenedor). Autocontenido: el agente no
> tiene contexto previo.

## Prerequisito (correr primero)

```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

Todos los tests con `gotest` (nunca `go test` a secas).

## Ya ejecutado (2026-07-10, LOCAL — verificar, no rehacer)

- Constante `ErrBrowserNotOpen` reemplazó los 10+ literales "Browser is not
  open... browser_open" (guard test: `mcp_errors_guard_test.go`).
- `models.go` migrado de los enums (`model.FieldText`...) a constructores
  Kind (`model.Text()`, `model.Int()`, `model.Bool()`); el repo compila
  contra `model v0.0.7` / `mcp v0.1.20`.

---

## 0. Diagnóstico (evidencia real, 2026-07-10)

`gotest ./...` tiene exactamente 2 tests rojos:

```
robust_interaction_test.go:78: Robust click failed: selector  character not allowed #
swipe_test.go:124: Swipe failed: selector  character not allowed #
```

Cadena verificada: `mcp.Request.Bind` (`mcp/tools.go:35`) →
`args.Validate(action)` (generado en `models_orm.go`) →
`model.ValidateFields` → `Field.Validate` → `Text().Validate("#btn")` →
la whitelist XSS de `model.Text()` no permite `#`.

El piso XSS de `model.Text()` es un DEFAULT para texto humano. Desde
`model v0.0.8` (contrato asentado en `model/docs/ARCHITECTURE.md` §8), un
Field que declara su propia whitelist POSITIVA en el `Permitted` embebido la
REEMPLAZA — el autor gobierna su charset y asume el encoding de salida. Los
campos de este repo cuyo contenido es un string **interpretado por el
browser** (selectores CSS, JS, URLs, valores arbitrarios a tipear) declaran
su whitelist explícita; regresión del contrato:
`model/tests/kind_permitted_override_test.go`.

## 1. Reglas de código (obligatorias)

- Strings repetidos → constantes tipadas; literales en lógica prohibidos.
- Errores se propagan; nada se traga en silencio.
- No tocar `models_orm.go` a mano (generado); si queda desalineado,
  regenerar con `ormc` — si el ormc fase B no está publicado, anotarlo en el
  resumen en vez de editarlo.

## 2. Etapa 1 — whitelists explícitas en los campos interpretados por el browser

Requiere `github.com/tinywasm/model` **≥ v0.0.8** en `go.mod` (bump +
`go mod tidy` si aún apunta a v0.0.7).

1. Definir en `models.go` tres charsets compartidos (regla §1 — nada de
   literales repetidos por campo):

```go
// Whitelists explícitas (model ≥ v0.0.8: reemplazan el piso default de
// Text() — ver model/docs/ARCHITECTURE.md §8). El encoding de salida es
// responsabilidad de quien renderice estos valores en HTML.
var (
	// permittedSelector: selectores CSS (#btn, .card > a[href^='x'],
	// div:nth-child(2n+1), [data-x="y"]).
	permittedSelector = model.Permitted{Letters: true, Numbers: true, Spaces: true,
		Extra: []rune(`#.-_[]()>~+*:,='"^$|`)}
	// permittedURL: RFC 3986 (unreserved + reserved + %).
	permittedURL = model.Permitted{Letters: true, Numbers: true,
		Extra: []rune(`:/?#[]@!$&'()*+,;=-._~%`)}
	// permittedFree: JS a evaluar, valores arbitrarios a tipear en inputs,
	// filtros de red — todo ASCII imprimible + saltos de línea/tab.
	permittedFree = model.Permitted{Letters: true, Numbers: true, Spaces: true,
		Tilde: true, BreakLine: true, Tab: true,
		Extra: []rune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")}
)
```

2. Asignarlos vía el `Permitted` embebido del Field (el `Type: model.Text()`
   NO cambia; `NotNull` se conserva donde ya está):

```go
{Name: "selector", Type: model.Text(), NotNull: true, Permitted: permittedSelector},
```

| Definition | Campo → whitelist |
|---|---|
| `ClickElementArgsModel` | `selector` → `permittedSelector` |
| `NavigateArgsModel` | `url` → `permittedURL` |
| `EmulateDeviceArgsModel` | `selector` → `permittedSelector` |
| `FillElementArgsModel` | `selector` → `permittedSelector`; `value` → `permittedFree` |
| `SwipeElementArgsModel` | `selector` → `permittedSelector` |
| `EvaluateJSArgsModel` | `script` → `permittedFree` |
| `GetNetworkLogsArgsModel` | `filter` → `permittedFree` |
| `GetSourceArgsModel` | `selector` → `permittedSelector` |
| `InspectElementArgsModel` | `selector` → `permittedSelector` |
| `GetStylesArgsModel` | `selector` → `permittedSelector` |
| `GetAssetArgsModel` | `url` → `permittedURL` |
| `InterceptRequestArgsModel` | `filter` → `permittedFree` |

Los demás campos (`mode`, `direction`, `type`, `action`, `port`, booleanos,
enteros) NO se tocan: sus valores son palabras simples que pasan el piso XSS
default de `Text()` y ese piso es deseable ahí.

**Verificación obligatoria:** `gotest ./...` — `TestRobustInteraction` y
`TestBrowserSwipe` deben pasar (usan `#btn` real vía chromedp). Todos los
tests verdes, no solo esos dos.

## 3. Etapa 2 — `tui.go`: DevBrowser pasa a `devtui.HandlerSelection` (sin gate)

Hoy `DevBrowser` en `tui.go` es un campo texto-libre: `Label()` dice
`"Auto Start Browser 't/f'"`, `Value()` devuelve `"t"`/`"f"`, y `Change`
tiene un default que **togglea con CUALQUIER input** (una key desconocida
cambia el estado — viola "nada se traga en silencio"). Auto-start es una
elección binaria excluyente: el contrato correcto es
`devtui.HandlerSelection` (radio/segmented-control), definido en
`devtui/interfaces.go`:

```go
type HandlerSelection interface {
    Name() string                 // identificador para logging
    Label() string                // etiqueta del grupo de botones
    Value() string                // KEY de la opción activa
    Change(newValue string)       // recibe la KEY confirmada
    Options() []map[string]string // pares ordenados {value: label}
}
```

Ejemplo de referencia: `devtui/example/HandlerSelection.go`
(`CompilerModeHandler` — nótese que sus shortcuts globales también entran
por `Change`). **Zero coupling:** la interfaz usa solo tipos stdlib —
devbrowser NO importa devtui; la satisface estructuralmente y
`AddHandler(handler any, ...)` hace el type-assert en devtui.

Cambios en `tui.go`:

1. Constantes tipadas para las keys (regla §1):

```go
const (
    autoStartOn  = "on"
    autoStartOff = "off"
    shortcutBrowserToggle = "B"
)
```

2. `Options()` devuelve `[{on: On}, {off: Off}]` (ordenadas).
3. `Value()` devuelve la KEY activa (`autoStartOn` si `h.AutoStart`).
4. `Change(newValue)` — switch explícito, se ELIMINA el default-toggle:
   - `autoStartOn` → `h.AutoStart = true`; `autoStartOff` → `false`; en
     ambos: `SaveConfig()`, `Logger(StatusMessage())`, `UI.RefreshUI()`.
   - `shortcutBrowserToggle` ("B") → conserva el comportamiento actual
     EXACTO (goroutine + guard `atomic.CompareAndSwapInt32` sobre `h.Busy`,
     open/close browser, return temprano sin RefreshUI).
   - key desconocida → `Logger` del error, sin cambiar estado (nada se
     traga en silencio).
5. `Label()` pasa a `"Auto Start"` (el grupo de botones ya muestra las
   opciones; el sufijo `'t/f'` desaparece).
6. `Shortcuts()` queda como está (`{"B": "toggle browser"}`) — devtui
   canaliza el shortcut por `Change("B")`, cubierto por el case del punto 4.
7. `StatusMessage()` actualiza su formato a las keys nuevas:
   `"Open | Auto-Start: on | Shortcut B"`.

Tests (con `gotest`): revisar los tests existentes de tui/config que
asserten `"t"`/`"f"` o el default-toggle y reescribirlos al contrato nuevo:

1. `Value()` inicial refleja `AutoStart`; `Change("off")` → `Value()=="off"`
   y `SaveConfig` persistió; `Change("on")` vuelve.
2. `Change("bogus")` → estado NO cambia y se loguea error.
3. `Options()` devuelve exactamente 2 entradas ordenadas `on`/`off`.
4. `Change("B")` sigue toggleando el browser (mock/flag) sin tocar
   `AutoStart`.

## 4. Criterios de aceptación

1. `gotest ./...` COMPLETO verde, incluidos `TestRobustInteraction` y
   `TestBrowserSwipe`.
2. `grep -c "permittedSelector\|permittedURL\|permittedFree" models.go` →
   los 3 vars + las 14 asignaciones de la tabla §2 (17 matches); cero
   `Permitted{` inline repetidos por campo.
3. `DevBrowser` satisface `HandlerSelection` sin importar devtui; el
   default-toggle desaparece (`Change` con key desconocida = error logueado,
   sin cambio de estado).
4. Shortcut "B" intacto.

## 5. Tabla de etapas

| # | Etapa | Archivos | Gate |
|---|-------|----------|------|
| 1 | Whitelists explícitas (Text + Permitted) | `models.go`, `go.mod` (model ≥ v0.0.8) | tests §2 |
| 2 | `DevBrowser` → HandlerSelection | `tui.go`, tests de tui | tests §3 |
