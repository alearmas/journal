[![CI](https://github.com/alearmas/tradingJournal/actions/workflows/ci.yml/badge.svg)](https://github.com/alearmas/tradingJournal/actions/workflows/ci.yml)

# Trading Journal – Usage Guide

This document explains **how to use all features** of the Trading Journal CLI.

---

## 1. Add a caución

Stores a new caución entry (JSON by default).

```bash
go run ./cmd/journal add \
  --principal 1000000.00 \
  --tna 85.5 \
  --term 1 \
  --fees 50.00 \
  --taxes 421.00 \
  --date 2026-01-10 \
  --notes "overnight balanz"
```

**What happens:**

* The caución is validated and calculated using `decimal.Decimal`
* Data is persisted to `data/cauciones.json` (unless configured otherwise)

---

## 2. List all cauciones

```bash
go run ./cmd/journal list
```

Outputs all stored cauciones in chronological order.

---

## 3. Monthly report

Summarizes all cauciones for a given month.

```bash
go run ./cmd/journal report --month 2026-01
```

**Includes:**

* Count of operations
* Total principal
* Gross interest
* Fees and taxes
* Net interest
* Weighted average TNA

---

## 4. Run tests

Financial calculations are covered by unit tests.

```bash
go test ./...
```

```bash
make test
make coverage
make coverage-check
```

This ensures:

* Deterministic results
* Correct rounding
* No floating-point drift

---

## 5. Export to CSV

Export all cauciones:

```bash
go run ./cmd/journal export --out cauciones.csv
```

Export a specific month:

```bash
go run ./cmd/journal export --out cauciones_2026-01.csv --month 2026-01
```

CSV can be opened directly in Excel / Google Sheets.

---

## 6. Compare caución vs PF vs MM

Compare expected returns for the same principal and time horizon.

```bash
go run ./cmd/journal compare \
  --principal 1000000.00 \
  --days 7 \
  --caucion-tna 85.5 \
  --fees 200.00 \
  --taxes 900.00 \
  --pf-tna 80.0 \
  --mm-tna 70.0
```

**Output:**

* Caución net return
* Plazo fijo gross return
* Money market gross return

---

## 7. Storage configuration

### Default (JSON)

```bash
# default
data/cauciones.json
```

### Custom JSON path

```bash
export JOURNAL_DATA=/path/to/cauciones.json
```

### SQLite (optional)

```bash
export JOURNAL_STORE=sqlite
export JOURNAL_DB=./data/journal.db
```

Then use the CLI normally:

```bash
go run ./cmd/journal add --principal 500000 --tna 82 --term 7
go run ./cmd/journal list
```

---

## Notes

* All monetary values use `decimal.Decimal`
* No floats are used in financial calculations
* JSON remains the default for transparency and auditability

---

This file can be linked or merged directly into `README.md`.
