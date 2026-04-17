[![CI](https://github.com/alearmas/tradingJournal/actions/workflows/ci.yml/badge.svg)](https://github.com/alearmas/tradingJournal/actions/workflows/ci.yml)

# Trading Journal

CLI y servidor HTTP en Go para registrar y analizar operaciones de **cauciones bursátiles** en el mercado argentino.
Todos los cálculos usan `decimal.Decimal` (sin `float64`) y la convención de **360 días por año**.

---

## Índice

- [Instalación](#instalación)
- [Modos de uso](#modos-de-uso)
- [CLI – Cauciones](#cli--cauciones)
- [CLI – Capital](#cli--capital)
- [Servidor HTTP](#servidor-http)
- [API REST](#api-rest)
- [Integración PPI](#integración-ppi)
- [Almacenamiento](#almacenamiento)
- [Variables de entorno](#variables-de-entorno)
- [Desarrollo](#desarrollo)
- [Notas técnicas](#notas-técnicas)

---

## Instalación

```bash
git clone https://github.com/alearmas/tradingJournal
cd tradingJournal/backend
go build -o journal ./cmd/journal   # CLI
go build -o server  ./cmd/server    # HTTP server
```

O con Make:

```bash
make server   # compila y corre el servidor
```

---

## Modos de uso

| Modo | Binario | Cuándo usarlo |
|------|---------|---------------|
| CLI interactivo | `./journal` | Carga manual, scripts, uso local rápido |
| Servidor HTTP | `./server` | Frontend, integraciones, Swagger UI |

---

## CLI – Cauciones

### Registrar una caución

```bash
journal add \
  --principal 1000000.00 \
  --tna 85.5 \
  --term 1 \
  --fees 50.00 \
  --taxes 421.00 \
  --date 2026-01-10 \
  --broker Balanz \
  --notes "overnight balanz"
```

| Flag | Descripción | Default |
|------|-------------|---------|
| `--principal` | Capital invertido | requerido |
| `--tna` | Tasa nominal anual (%) | requerido |
| `--term` | Plazo en días | requerido |
| `--fees` | Comisiones | 0 |
| `--taxes` | Retenciones/impuestos | 0 |
| `--date` | Fecha de operación `YYYY-MM-DD` | hoy |
| `--broker` | Nombre del broker | Balanz |
| `--notes` | Notas libres | — |

**Cálculo aplicado:**
```
Interés bruto = Principal × (TNA/100) × (días/360)
Interés neto  = Bruto − Comisiones − Retenciones
```

### Otros comandos

```bash
journal list                          # listar todas las cauciones
journal summary                       # resumen global (totales)
journal report --month 2026-01        # reporte mensual con TNA ponderada
journal export --out cauciones.csv    # exportar a CSV
journal export --out jan.csv --month 2026-01

journal compare \
  --principal 1000000.00 \
  --days 7 \
  --caucion-tna 85.5 \
  --fees 200.00 \
  --taxes 900.00 \
  --pf-tna 80.0 \
  --mm-tna 70.0
```

El comando `compare` muestra el retorno neto de la caución frente a Plazo Fijo y Money Market para el mismo capital y plazo.

---

## CLI – Capital

Registrá movimientos de dinero para rastrear capital disponible, porcentaje desplegado y P&L total.

```bash
journal deposit  --broker Balanz --amount 1000000.00 --date 2026-01-01
journal withdraw --broker Balanz --amount 200000.00  --date 2026-02-15

journal balance             # todos los brokers
journal balance --broker PPI
```

**Ejemplo de salida:**
```
--- Balanz ---
  Depositado:   1.000.000,00
  Retirado:       200.000,00
  Saldo neto:     800.000,00
  Desplegado:     600.000,00   ← cauciones activas (vencen en el futuro)
  Disponible:     200.000,00
  Ganancia:         5.234,72
  P&L %:              0.52%
```

> **Desplegado** = suma del principal de cauciones cuya fecha de vencimiento es posterior a hoy.
> Cuando una caución vence, el capital vuelve automáticamente al disponible sin necesidad de registrar ningún evento extra.

---

## Servidor HTTP

```bash
./server
# o
make server
```

Por defecto escucha en `:8080`. Cambiar con `JOURNAL_ADDR`.

### Swagger UI

Una vez levantado el servidor, la documentación interactiva está en:

```
http://localhost:8080/swagger/index.html
```

Para regenerar los docs después de modificar handlers:

```bash
make swag
```

---

## API REST

Base path: `/api/v1`

### Cauciones

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `POST` | `/cauciones` | Crear caución |
| `GET` | `/cauciones` | Listar cauciones |
| `GET` | `/cauciones/summary` | Resumen global |
| `GET` | `/cauciones/report?month=YYYY-MM` | Reporte mensual |
| `GET` | `/cauciones/export?month=YYYY-MM` | Descargar CSV |

### Movimientos de capital

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `POST` | `/movimientos/deposit` | Registrar depósito |
| `POST` | `/movimientos/withdraw` | Registrar retiro |
| `GET` | `/movimientos` | Listar movimientos |

### Capital y comparación

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `GET` | `/capital/balance?broker=X` | Resumen de capital por broker |
| `GET` | `/compare` | Comparar instrumentos |

### Sync PPI

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| `POST` | `/sync?date_from=YYYY-MM-DD&date_to=YYYY-MM-DD` | Importar desde PPI |
| `GET` | `/ppi/movements` | Movimientos raw PPI (debug) |
| `GET` | `/ppi/orders` | Órdenes raw PPI (debug) |

---

## Integración PPI

El servidor puede sincronizar automáticamente movimientos y cauciones desde la API de PPI (Prima Portafolio de Inversiones).

```bash
# Configurar credenciales
export PPI_PUBLIC_KEY=tu_api_key
export PPI_PRIVATE_KEY=tu_api_secret
export AUTHORIZED_CLIENT=tu_client_id
export CLIENT_KEY=tu_client_key
export PPI_ACCOUNT=tu_numero_de_cuenta
export PPI_SANDBOX=false   # true para pruebas

# Sincronizar últimos 30 días (default)
curl -X POST http://localhost:8080/api/v1/sync

# Rango específico
curl -X POST "http://localhost:8080/api/v1/sync?date_from=2026-01-01&date_to=2026-03-31"
```

**Comportamiento del sync:**
- Deduplica por fecha + tipo + monto (movimientos) o fecha + plazo + principal (cauciones)
- Clasifica descripciones de PPI como `deposito`/`retiro` automáticamente
- Parsea el ticker de cada orden para extraer el plazo (ej: `"PESOS7"` → 7 días)
- Devuelve conteo de importados, omitidos y errores por categoría

---

## Almacenamiento

### JSON (default)

Archivos legibles por humanos, fáciles de versionar con git:

```
data/cauciones.json
data/movimientos.json
```

### SQLite (opcional)

```bash
export JOURNAL_STORE=sqlite
export JOURNAL_DB=./data/journal.db

journal add --principal 500000 --tna 82 --term 7
```

Ambos backends implementan las mismas interfaces (`CaucionRepository`, `MovimientoRepository`) — se puede cambiar sin tocar la lógica de negocio.

---

## Variables de entorno

| Variable | Default | Descripción |
|----------|---------|-------------|
| `JOURNAL_STORE` | `json` | Backend: `json` o `sqlite` |
| `JOURNAL_DATA` | `data/cauciones.json` | Path del JSON de cauciones |
| `JOURNAL_MOVEMENTS` | `data/movimientos.json` | Path del JSON de movimientos |
| `JOURNAL_DB` | `data/journal.db` | Path del SQLite |
| `JOURNAL_ADDR` | `:8080` | Dirección del servidor HTTP |
| `PPI_PUBLIC_KEY` | — | API key de PPI |
| `PPI_PRIVATE_KEY` | — | API secret de PPI |
| `PPI_SANDBOX` | `true` | Usar sandbox de PPI |
| `AUTHORIZED_CLIENT` | — | Client ID autorizado de PPI |
| `CLIENT_KEY` | — | Client key de PPI |
| `PPI_ACCOUNT` | — | Número de cuenta PPI para sync |

---

## Desarrollo

```bash
make test            # suite completo + race detector
make race            # detección de race conditions
make coverage        # reporte HTML (coverage/coverage.html)
make coverage-check  # verifica umbral mínimo del 70%
make fmt             # formatear código
make lint            # golangci-lint
make install-hooks   # git pre-commit hooks
make swag            # regenerar docs de Swagger
make server          # correr el servidor HTTP
```

---

## Notas técnicas

- **Sin `float64`** en ningún cálculo financiero — todo usa `shopspring/decimal`
- **360 días/año** — convención estándar del mercado de renta fija
- **Arquitectura hexagonal** — dominio, puertos e implementaciones desacoplados
- **JSON como default** para transparencia y auditabilidad (legible por humanos, versionable con git)
- **SQLite** para uso más intensivo — mismo motor, tablas por entidad, decimales guardados como TEXT
- **Thread-safe** — mutex en repos JSON, operaciones context-aware en SQLite
- **Swagger** generado desde anotaciones en código — `make swag` luego de cambiar handlers
