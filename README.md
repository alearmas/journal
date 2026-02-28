[![CI](https://github.com/alearmas/tradingJournal/actions/workflows/ci.yml/badge.svg)](https://github.com/alearmas/tradingJournal/actions/workflows/ci.yml)

# Trading Journal – Guía de uso

CLI en Go para registrar y analizar operaciones de **cauciones bursátiles** en el mercado argentino.
Todos los cálculos usan `decimal.Decimal` (sin `float64`) y la convención de **360 días por año**.

---

## Instalación rápida

```bash
git clone https://github.com/alearmas/tradingJournal
cd tradingJournal
go build -o journal ./cmd/journal
./journal --help
```

---

## Comandos de cauciones

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
|---|---|---|
| `--principal` | Capital invertido | requerido |
| `--tna` | Tasa nominal anual (%) | requerido |
| `--term` | Plazo en días | requerido |
| `--fees` | Comisiones | 0 |
| `--taxes` | Retenciones/impuestos | 0 |
| `--date` | Fecha de la operación `YYYY-MM-DD` | hoy |
| `--broker` | Nombre del broker | Balanz |
| `--notes` | Notas libres | — |

**Cálculo aplicado:**
```
Interés bruto = Principal × (TNA/100) × (días/360)
Interés neto  = Bruto − Comisiones − Retenciones
```

### Listar todas las cauciones

```bash
journal list
```

### Resumen global

```bash
journal summary
```

Muestra totales de principal, interés bruto/neto, comisiones e impuestos de todas las operaciones.

### Reporte mensual

```bash
journal report --month 2026-01
```

Incluye: conteo, principal total, intereses, comisiones, impuestos y **TNA promedio ponderada**.

### Exportar a CSV

```bash
# Todas las cauciones
journal export --out cauciones.csv

# Solo un mes
journal export --out cauciones_2026-01.csv --month 2026-01
```

Compatible con Excel / Google Sheets.

### Comparar instrumentos

Compara el retorno neto de una caución contra Plazo Fijo y Money Market para el mismo capital y plazo.

```bash
journal compare \
  --principal 1000000.00 \
  --days 7 \
  --caucion-tna 85.5 \
  --fees 200.00 \
  --taxes 900.00 \
  --pf-tna 80.0 \
  --mm-tna 70.0
```

---

## Comandos de capital

Registrá los movimientos de dinero en cada cuenta de broker para rastrear tu capital disponible, el porcentaje desplegado y tu P&L total.

### Depositar capital

```bash
journal deposit \
  --broker Balanz \
  --amount 1000000.00 \
  --date 2026-01-01 \
  --notes "transferencia inicial"
```

### Retirar capital

```bash
journal withdraw \
  --broker Balanz \
  --amount 200000.00 \
  --date 2026-02-15
```

| Flag | Descripción | Default |
|---|---|---|
| `--broker` | Nombre del broker | Balanz |
| `--amount` | Monto (siempre positivo) | requerido |
| `--date` | Fecha `YYYY-MM-DD` | hoy |
| `--notes` | Notas libres | — |

### Ver balance por broker

```bash
# Todos los brokers
journal balance

# Filtrar por broker
journal balance --broker Balanz
```

**Ejemplo de salida:**
```
--- Balanz ---
  Depositado:  1000000.00
  Retirado:       200000.00
  Saldo neto:     800000.00
  Desplegado:     600000.00   ← cauciones activas (vencen en el futuro)
  Disponible:     200000.00
  Ganancia:         5234.72
  P&L %:             0.6543%
```

> **Desplegado** = suma del principal de cauciones cuya fecha de vencimiento es posterior a hoy.
> Cuando una caución vence, el dinero vuelve automáticamente al disponible — no se necesita registrar ningún evento extra.

---

## Configuración de almacenamiento

### JSON (default)

```bash
# Archivos por defecto
data/cauciones.json
data/movimientos.json
```

### Variables de entorno

| Variable | Default | Descripción |
|---|---|---|
| `JOURNAL_STORE` | `json` | Backend: `json` o `sqlite` |
| `JOURNAL_DATA` | `data/cauciones.json` | Path del JSON de cauciones |
| `JOURNAL_MOVEMENTS` | `data/movimientos.json` | Path del JSON de movimientos |
| `JOURNAL_DB` | `data/journal.db` | Path del SQLite |

### SQLite (opcional)

```bash
export JOURNAL_STORE=sqlite
export JOURNAL_DB=./data/journal.db

journal add --principal 500000 --tna 82 --term 7
journal deposit --amount 500000 --broker Balanz
journal balance
```

---

## Tests

```bash
make test           # suite completo + race detector
make coverage       # reporte HTML de cobertura
make coverage-check # verifica umbral mínimo del 70%
```

---

## Notas técnicas

- Sin `float64` en ningún cálculo financiero — todo usa `shopspring/decimal`
- JSON como default para transparencia y auditabilidad (legible por humanos, fácil de versionar con git)
- SQLite para uso más intensivo — mismo archivo, tablas separadas por entidad
