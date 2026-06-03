# PLAN: browser_position no se guarda/aplica (browser_size sí) — REVISIÓN

## Estado: PARA REVISIÓN (no se ha tocado código)

## Síntoma

- Al mover el navegador con el mouse, `browser_position` NO se actualiza en `.env`.
- Al redimensionarlo, `browser_size` SÍ se actualiza.
- "Antes funcionaba"; ahora no.

## Reproducción (test real, solo chromedp — cross-platform)

Archivo: [position_repro_test.go](../position_repro_test.go) (paquete interno,
lanza Chrome headful real, sin herramientas de SO).

### Test 1 — `TestReproSetWindowBoundsIgnored` (automático, FALLA en `gotest`)

Corre automáticamente cuando hay display (se salta solo en CI sin pantalla). NO
necesita variable de entorno: aparece como `FAIL` en una corrida normal.

Lanza el navegador EXACTAMENTE como la app (`CreateBrowserContext`) y pide a CDP
mover+redimensionar con una sola llamada `SetWindowBounds(Left=450, Top=250,
Width=1000, Height=650)`. Lee de vuelta con `GetWindowForTarget`:

| Atributo | Pedido | Resultado CDP | Veredicto |
|---|---|---|---|
| Tamaño   | 1000×650 | **1000×651** | ✅ aplicado (±1px por decoración) |
| Posición | 450,250  | **0,0**      | ❌ IGNORADO |

**Hallazgo central, reproducible:** la MISMA llamada CDP aplica el tamaño y
descarta la posición. La asimetría NO está en nuestro código (posición y tamaño
se guardan en el mismo `checkAndSaveGeometry`, leyendo el mismo `bounds`), sino
en que Chrome/el gestor de ventanas **honra peticiones de tamaño e ignora
peticiones de posición**.

### Test 2 — `TestReproManualMoveDetected` (manual, para confirmar el lado "guardar")

Abre el navegador y hace polling de la geometría 1×/seg durante 20 s. El revisor
debe **arrastrar y redimensionar la ventana con el mouse**. El log indica si CDP
detecta el movimiento manual (posición) y el redimensionado manual (tamaño).

Ejecutar:
```
TW_REPRO=1 go test ./ -run TestReproManualMoveDetected -v -timeout 60s
```

Esto responde la pregunta abierta: ¿`GetWindowForTarget` reporta el movimiento
manual de la ventana? Si SÍ → el guardado funciona una vez la posición es
correcta. Si NO → este WM tampoco expone la posición tras un arrastre manual, y
no hay forma cross-platform por CDP de leerla.

## Por qué tamaño funciona y posición no

| | Aplicar al abrir | Guardar al cambiar |
|---|---|---|
| **Tamaño** | `--window-size` (flag de arranque) + `SetWindowBounds` → ambos honrados | monitor lee `bounds.Width/Height` (cambia con resize) → guarda ✓ |
| **Posición** | `--window-position` (flag) IGNORADO + `SetWindowBounds` IGNORADO | monitor lee `bounds.Left/Top`; si el WM no los actualiza, `newPos == b.Position` → nunca guarda |

Notas:
- `MonitorWidth/MonitorHeight` (detección de pantalla) NO interviene: solo fija el
  tamaño inicial cuando no hay `browser_size`. No toca la posición.
- El primer diagnóstico (con flags distintos) mostró `SetWindowBounds` aplicando
  posición OK; con los flags reales de la app, NO. Diferencia de flags pendiente
  de aislar (ver "Investigación pendiente").

## Causa raíz (hipótesis a validar con Test 2)

El gestor de ventanas del entorno **ignora el posicionamiento programático** de la
ventana (tanto el flag `--window-position` como la llamada CDP `SetWindowBounds`),
mientras **respeta el dimensionamiento**. Por eso:

1. Al abrir, la ventana no va a la posición configurada (queda en 0,0).
2. `b.Position` arranca en `"0,0"`; si el WM no reporta movimientos, CDP devuelve
   `0,0`, `newPosition == b.Position`, y nunca se guarda.

Esto es coherente con "el tamaño se guarda y la posición no" SIN que haya una
asimetría real en nuestro código de guardado.

## Investigación pendiente (antes de proponer fix definitivo)

1. **Ejecutar Test 2** y observar si el arrastre manual cambia `Left/Top` en CDP.
   - Si cambia → el bug es solo de APLICAR al abrir; el guardado ya sirve. Fix:
     mantener `b.Position` y dejar que el monitor guarde (quitar el guard
     `x>0||y>0` que añadí, para reflejar exactamente la ruta de tamaño).
   - Si NO cambia → este WM no expone la posición; no hay solución cross-platform
     vía CDP. Se documenta la limitación y se busca alternativa.
2. **Aislar la diferencia de flags** entre el primer diagnóstico (SetWindowBounds
   aplicó posición) y `CreateBrowserContext` (no aplicó). Candidatos a probar
   quitando uno por uno: `disable-blink-features`, `use-fake-ui-for-media-stream`,
   `disable-cache`, ausencia de `--window-position`.

## Propuesta de fix (PARA DISCUSIÓN, no implementada)

Reflejar la ruta de `browser_size` lo más posible:

1. **Aplicar al abrir:** restaurar el intento por flag de arranque
   `chromedp.Flag("window-position", h.Position)` (espejo de `WindowSize`), que es
   cross-platform y honrado en Windows/Mac y muchos WM de Linux; complementarlo con
   `SetWindowBounds` tras `ReadyChan` como respaldo. En WMs que ignoran ambos, no
   hay más que se pueda hacer vía CDP.
2. **Guardar al mover:** quitar el guard `x>0||y>0` para que posición sea idéntico
   a tamaño (guardar cuando `newPosition != b.Position`).
3. Confirmar con Test 2 que el monitor capta el arrastre manual.

## Archivos involucrados (referencia, sin cambios aún)

| Archivo | Rol |
|---|---|
| [context.go](../context.go) | flags de arranque (`WindowSize` ok; `window-position` fue removido) |
| [position.go](../position.go) | `applyConfiguredPosition` (usa `SetWindowBounds`, hoy ignorado) |
| [monitor_geometry.go](../monitor_geometry.go) | `checkAndSaveGeometry`/`SaveGeometry` (guardado; guard a revisar) |

## Tests entregados

- [position_repro_test.go](../position_repro_test.go)
  - `TestReproSetWindowBoundsIgnored` — automático (corre con display, FALLA en
    `gotest`), reproduce que la posición programática es ignorada mientras el
    tamaño se aplica.
  - `TestReproManualMoveDetected` — manual (gated por `TW_REPRO=1`), para
    confirmar si CDP capta el arrastre manual con el mouse.
