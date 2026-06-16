# devbrowser — PLAN: Navegación relativa + reporte de ruta actual

> Estado: Borrador para revisión · Objetivo: permite al LLM navegar con un path relativo
> a la app en ejecución (sin adivinar puerto) y siempre conoce su ruta actual.
>
> ⚠️ Prescriptivo. Ver §5 (Invariantes) y §6 (Aceptación).

---

## 1. Problema (verificado en código)

`browser_navigate` (`devbrowser/mcp-navigation.go:13`) vincula `NavigateArgs{ URL string }`
(`devbrowser/models.go:16`) y llama `NavigateToURL(args.URL)` (`devbrowser/devbrowser.go:184`),
que ejecuta `chromedp.Navigate(url)`. El esquema requiere una URL completa, así que el LLM
debe saber el host:puerto — que ninguna tool reporta. La cadena de resultado es solo
`"Navigated to <url>"`; no confirma dónde realmente aterrizó la página.

Estado existente relevante:

- `DevBrowser` ya almacena `LastPort string` y `LastHttps bool`
  (`devbrowser/devbrowser.go:36-37`), establecidos cuando el navegador se abre
  (`OpenBrowser(port, https)`).
- La URL actual es obtendible vía CDP: `chromedp.Location(&res.PageURL)` ya se usa para
  screenshots (`devbrowser/screenshot_utils.go:49`, `ScreenshotResult.PageURL` en `:13`).

Así que tanto la URL base como la ruta actual son derivables dentro de devbrowser — sin
inyección de app necesaria.

## 2. Objetivo

- `browser_navigate` acepta una URL absoluta O un path relativo a la app en ejecución
  (ej. `/login`), resuelto contra `LastPort`/`LastHttps`.
- El resultado de navigate reporta la URL absoluta resuelta Y la ruta actual después de la navegación.

## 3. Diseño

### 3.1 Resolución de path relativo

En `browser_navigate.Execute` (`devbrowser/mcp-navigation.go`):

- If `args.URL` has a scheme (`http://`, `https://`) → use as-is.
- Else (starts with `/` or is relative) → build base from state:
  `scheme := "http"; if b.LastHttps { scheme = "https" }`,
  `base := scheme + "://localhost:" + b.LastPort`, then resolve `args.URL` against `base`
  (use `net/url` `ResolveReference` for correctness with `/` and relative forms).
- If `b.LastPort == ""` (browser not opened yet) → return a clear error:
  `"browser has no active app port; open the app first"`.

Mantén la guardia `IsOpenFlag` existente (`devbrowser/mcp-navigation.go:19`).

### 3.2 Ruta actual en el resultado

Después de una navegación exitosa, lee la ubicación actual e inclúyela:

- Add a helper `func (b *DevBrowser) CurrentURL() (string, error)` running
  `chromedp.Location(&u)` against `b.Ctx` (mirrors `screenshot_utils.go:49`).
- Result text: `"Navigated to <resolved> (current: <CurrentURL>)"`.

Schema description updated (shape unchanged):

```json
{ "url": "string — absolute URL or path relative to the running app, e.g. /login" }
```

### 3.3 Actualiza la descripción equivalente anunciada por el daemon

Porque el daemon actualmente refleja `browser_navigate` (`app/daemon.go:444`), una vez que
el plan de app elimine el reflejo codificado, la descripción de devbrowser aquí (aquí) se
vuelve la fuente única. Asegura que la descripción declare el comportamiento de path relativo.

## 4. Pasos

1. Añade helper `CurrentURL()` (`devbrowser/devbrowser.go`, cerca de `NavigateToURL:184`).
2. Resuelve paths relativos en `browser_navigate.Execute` usando `LastPort`/`LastHttps`.
3. Incluye URL resuelta + ruta actual en el resultado; actualiza la descripción del esquema
   en el doc de texto `EncodeSchema(new(NavigateArgs))`.

## 5. Invariantes / prohibiciones

- **No** añadas un campo/inyección de URL base desde app; usa `LastPort`/`LastHttps` existentes.
- **No** cambies el conjunto de campos de `NavigateArgs` (`devbrowser/models.go:16`); solo
  la semántica + descripción cambian.
- "Listar rutas disponibles" es propiedad de `app_info` (plan de app), NO aquí — devbrowser
  reporta solo la ruta **actual**.

## 6. Aceptación

- `browser_navigate("/")` y `browser_navigate("/login")` funcionan sin que el LLM sepa el puerto.
- Una URL absoluta sigue funcionando sin cambios.
- El resultado establece la ruta absoluta actual después de la navegación.
- `go build ./... && go test ./...` en verde para devbrowser.

## 7. Tests

Añade a los tests de devbrowser:

1. **Resolución:** con `LastPort="6060"`, `LastHttps=false`, aserta que `/login` se resuelve a
   `http://localhost:6060/login` y una URL absoluta pasa a través sin cambios (pure helper de
   resolución, sin navegador real).
2. **Sin puerto:** con `LastPort=""`, aserta el error claro.
3. (Si es factible con el harness de test de navegador existente) la ruta actual se incluye en
   el resultado de navigate.

## 8. Actualizaciones de documentación

- Actualiza los docs de herramienta MCP de devbrowser para establecer que `browser_navigate`
  acepta paths relativos y reporta la ruta actual.
