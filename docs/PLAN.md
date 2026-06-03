# PLAN: browser_position no se aplica ni se guarda

## Estado: PENDIENTE DE REVISIÓN

## Síntoma reportado

1. Al abrir el navegador NO se aplica la posición configurada en `.env`
   (`browser_position`). La ventana siempre aparece en `0,0`.
2. Al mover el navegador manualmente, la nueva posición NO se guarda en `.env`.
3. El tamaño (`browser_size`) SÍ funciona correctamente en ambos sentidos:
   se aplica al abrir y se guarda al redimensionar.

`.env` de ejemplo (no cambia nunca el valor de posición):
```
dev_mode=true
browser_size=900,700
browser_position=0,0
```

## Evidencia empírica (diagnóstico con Chrome real headful)

Se lanzó Chrome con `--window-position=50,50` y `--window-size=800,600`,
luego se leyó la geometría con CDP `GetWindowForTarget`, después se movió la
ventana con CDP `SetWindowBounds(300,200)` y se volvió a leer:

| Momento | Posición leída por CDP | Tamaño leído por CDP |
|---|---|---|
| Tras lanzar (`--window-position=50,50`) | **Left=0, Top=0** ❌ flag ignorado | Width=800, Height=600 ✓ |
| Tras `SetWindowBounds(300,200)` | **Left=300, Top=200** ✓ | Width=800, Height=600 ✓ |

Entorno: `DISPLAY=:0` y `WAYLAND_DISPLAY=wayland-0` (sesión Wayland; Chrome corre
bajo XWayland). El mismo comportamiento aplica a cualquier WM que ignore el flag.

### Conclusiones de la evidencia

- **`chromedp` SÍ lee la posición correctamente** (devuelve 300,200 tras moverla).
  La lectura nunca fue el problema.
- **El flag de arranque `--window-position` es IGNORADO** por el compositor: la
  ventana arranca en `0,0` sin importar el valor del flag.
- **`SetWindowBounds` (llamada CDP programática) SÍ mueve la ventana** y la
  posición queda legible.
- **`--window-size` SÍ es respetado** por el compositor (de ahí que el tamaño
  funcione y la posición no).

## Causa raíz

El código depende del flag de línea de comandos `--window-position` para colocar
la ventana al abrir ([context.go](../context.go) línea 16):

```go
chromedp.Flag("window-position", h.Position),   // ← IGNORADO por el compositor
chromedp.WindowSize(h.Width, h.Height),          // ← SÍ respetado
```

Secuencia del fallo:

1. Al abrir, el flag `--window-position=0,0` se ignora → la ventana queda en `0,0`
   de todas formas (coincidencia con el valor por defecto, oculta el bug).
2. El monitor (`checkAndSaveGeometry`, cada 2 s) lee con CDP: `x=0, y=0`.
3. `newPosition="0,0"` es igual a `b.Position="0,0"` → **nunca se guarda**.
4. Si el usuario configura otra posición (ej. `1930,0` para segundo monitor), al
   abrir el flag se ignora, la ventana aparece en `0,0`, y el monitor entonces
   **sobrescribe** `1930,0` con `0,0`.

El tamaño no sufre esto porque `--window-size` SÍ se aplica: la ventana arranca
con el tamaño configurado, CDP lo reporta igual, no hay sobrescritura, y los
cambios manuales de tamaño se leen y guardan bien.

NOTA: la detección de monitor (`MonitorWidth/MonitorHeight`) NO interviene aquí.
Solo afecta el tamaño inicial cuando no hay `browser_size` configurado. No toca
la posición. No es parte de este bug.

## Solución propuesta

### Fix 1 — Aplicar la posición vía CDP `SetWindowBounds` tras abrir (PRINCIPAL)

Como el flag `--window-position` es ignorado pero `SetWindowBounds` funciona,
forzar la posición programáticamente después de que el navegador esté listo.

En [OpenBrowser.go](../OpenBrowser.go), tras recibir `ReadyChan`:
```go
go h.applyConfiguredPosition()   // aplica b.Position vía SetWindowBounds CDP
go h.monitorBrowserGeometry()
```

Nuevo método en [position.go](../position.go):
```go
func (b *DevBrowser) applyConfiguredPosition() {
    // parse b.Position ("x,y") -> x, y
    // GetWindowForTarget -> windowID
    // SetWindowBounds(windowID, Bounds{Left:x, Top:y, WindowState: normal})
}
```

Efecto: la ventana queda en la posición configurada; a partir de ahí
`GetWindowForTarget` reporta esa posición y el monitor detecta correctamente los
movimientos posteriores del usuario y los guarda.

CUIDADO con `omitempty`: el struct `browser.Bounds` marca `Left` y `Top` como
`json:"left,omitempty"`. Enviar `Left:0, Top:0` los descarta del JSON → no mueve.
Para el caso `0,0` no es problema (la ventana ya arranca en 0,0). Para posiciones
no-cero funciona bien.

### Fix 2 — Guardar posición junto con el tamaño (consistencia)

En `checkAndSaveGeometry`/`SaveGeometry`, cuando se guarda el tamaño, guardar
también la posición en la misma operación, para que ambas claves del `.env`
queden siempre sincronizadas aunque la posición no haya cambiado respecto a
`b.Position`.

### Fix 3 — Guard `x > 0 || y > 0` como cinturón de seguridad (DECIDIDO: mantener)

El guard evita que un `0,0` transitorio (reportado justo tras el arranque, antes
de que `applyConfiguredPosition` mueva la ventana, o si `SetWindowBounds`
fallara) sobrescriba una posición configurada no-cero. Como la evidencia muestra
que CDP lee coordenadas reales una vez colocada la ventana, el movimiento normal
del usuario se sigue registrando. Se mantiene por robustez; no cripplea la
lectura de posición.

## Archivos a cambiar

| Archivo | Cambio |
|---|---|
| [position.go](../position.go) | añadir `applyConfiguredPosition()` con `SetWindowBounds` CDP |
| [OpenBrowser.go](../OpenBrowser.go) | llamar `applyConfiguredPosition()` tras `ReadyChan` |
| [monitor_geometry.go](../monitor_geometry.go) | guardar posición junto al tamaño; mantener guard `x>0\|\|y>0` |

## Tests

- [tests/geometry_monitor_save_test.go](../tests/geometry_monitor_save_test.go)
  - `TestPositionNotOverwrittenWhenCDPReportsZero` — un `0,0` espurio no
    sobrescribe una posición configurada.
  - `TestPositionSavedWhenSizeChanges` — al cambiar el tamaño, la posición se
    persiste en la misma operación.
- [tests/geometry_position_tracking_test.go](../tests/geometry_position_tracking_test.go)
  - `TestPositionSavedOnUserMovement` — cuando CDP reporta coordenadas reales
    (caso normal confirmado por el diagnóstico), el movimiento del usuario se
    guarda.
  - `TestSaveGeometryWriteCount` — verifica que los valores quedan correctos tras
    un cambio simultáneo de posición y tamaño.
- Diagnóstico manual (headful, no commiteado): confirmó que `--window-position`
  se ignora y que `SetWindowBounds` funciona. Se puede reintroducir como test
  gated por variable de entorno si se quiere repetir la verificación.

## Verificación manual tras el fix

1. Poner `browser_position=400,200` en `.env`, abrir → la ventana debe aparecer
   en 400,200 (no en 0,0).
2. Mover la ventana a otra posición → `.env` debe actualizarse al valor nuevo.
3. Cerrar y reabrir → debe reaparecer en la última posición guardada.
