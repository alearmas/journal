package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary compiles the CLI binary into a temp directory and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "journal")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestCLI_NoArgs(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit error with no args")
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Fatalf("expected usage output, got: %s", out)
	}
}

func TestCLI_UnknownCommand(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "unknown")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit error for unknown command")
	}
	if !strings.Contains(string(out), "Usage:") {
		t.Fatalf("expected usage output, got: %s", out)
	}
}

func TestCLI_InvalidStore(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "list")
	cmd.Env = append(os.Environ(), "JOURNAL_STORE=invalid")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected exit error for invalid store")
	}
	if !strings.Contains(string(out), "invalid JOURNAL_STORE") {
		t.Fatalf("expected invalid store error, got: %s", out)
	}
}

func TestCLI_ListEmpty_JSON(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")

	cmd := exec.Command(bin, "list")
	cmd.Env = append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "No cauciones found") {
		t.Fatalf("expected 'No cauciones found', got: %s", out)
	}
}

func TestCLI_AddAndList_JSON(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")
	env := append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)

	// Add a caucion
	addCmd := exec.Command(bin, "add",
		"--principal", "1000000",
		"--tna", "85.5",
		"--term", "1",
		"--fees", "50",
		"--taxes", "421",
		"--date", "2026-01-10",
		"--notes", "integration test",
	)
	addCmd.Env = env
	out, err := addCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("add failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Saved caucion") {
		t.Fatalf("expected 'Saved caucion', got: %s", out)
	}

	// List should show the caucion
	listCmd := exec.Command(bin, "list")
	listCmd.Env = env
	out, err = listCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("list failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "integration test") {
		t.Fatalf("expected caucion in list, got: %s", out)
	}
}

func TestCLI_Summary_JSON(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")
	env := append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)

	// Add a caucion first
	addCmd := exec.Command(bin, "add",
		"--principal", "1000000",
		"--tna", "85.5",
		"--term", "1",
		"--fees", "50",
		"--taxes", "421",
		"--date", "2026-01-10",
	)
	addCmd.Env = env
	if out, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("add: %v\n%s", err, out)
	}

	// Summary
	sumCmd := exec.Command(bin, "summary")
	sumCmd.Env = env
	out, err := sumCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("summary: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Count: 1") {
		t.Fatalf("expected Count: 1, got: %s", out)
	}
	if !strings.Contains(string(out), "1000000.00") {
		t.Fatalf("expected principal in summary, got: %s", out)
	}
}

func TestCLI_Report_JSON(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")
	env := append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)

	// Add
	addCmd := exec.Command(bin, "add",
		"--principal", "1000000",
		"--tna", "85.5",
		"--term", "1",
		"--date", "2026-01-10",
	)
	addCmd.Env = env
	if out, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("add: %v\n%s", err, out)
	}

	// Report
	repCmd := exec.Command(bin, "report", "--month", "2026-01")
	repCmd.Env = env
	out, err := repCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("report: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Month: 2026-01") {
		t.Fatalf("expected month in report, got: %s", out)
	}
}

func TestCLI_ReportMissingMonth(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")

	cmd := exec.Command(bin, "report")
	cmd.Env = append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for missing --month")
	}
	if !strings.Contains(string(out), "missing --month") {
		t.Fatalf("expected missing month error, got: %s", out)
	}
}

func TestCLI_Export_JSON(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")
	csvPath := filepath.Join(tmp, "export.csv")
	env := append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)

	// Add
	addCmd := exec.Command(bin, "add",
		"--principal", "1000000",
		"--tna", "85.5",
		"--term", "1",
		"--date", "2026-01-10",
	)
	addCmd.Env = env
	if out, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("add: %v\n%s", err, out)
	}

	// Export
	expCmd := exec.Command(bin, "export", "--out", csvPath)
	expCmd.Env = env
	out, err := expCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "CSV exported to:") {
		t.Fatalf("expected export success, got: %s", out)
	}

	// Verify CSV file exists
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Fatal("CSV file was not created")
	}
}

func TestCLI_Compare(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")

	cmd := exec.Command(bin, "compare",
		"--principal", "1000000",
		"--days", "1",
		"--caucion-tna", "85.5",
		"--fees", "50",
		"--taxes", "421",
		"--pf-tna", "80",
		"--mm-tna", "70",
	)
	cmd.Env = append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compare: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Caucion net:") {
		t.Fatalf("expected compare output, got: %s", out)
	}
	if !strings.Contains(string(out), "1904.00") {
		t.Fatalf("expected caucion net 1904.00, got: %s", out)
	}
}

func TestCLI_AddValidationError(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")

	// Missing required term (defaults to 0, should error)
	cmd := exec.Command(bin, "add",
		"--principal", "1000",
		"--tna", "85",
	)
	cmd.Env = append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for zero term")
	}
	if !strings.Contains(string(out), "error:") {
		t.Fatalf("expected error message, got: %s", out)
	}
}

func TestCLI_AddInvalidDecimal(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dataPath := filepath.Join(tmp, "data.json")

	cmd := exec.Command(bin, "add",
		"--principal", "not-a-number",
		"--tna", "85",
		"--term", "1",
	)
	cmd.Env = append(os.Environ(),
		"JOURNAL_STORE=json",
		"JOURNAL_DATA="+dataPath,
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error for invalid decimal")
	}
	if !strings.Contains(string(out), "invalid --principal") {
		t.Fatalf("expected invalid principal error, got: %s", out)
	}
}

func TestCLI_AddAndList_SQLite(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "journal.db")
	env := append(os.Environ(),
		"JOURNAL_STORE=sqlite",
		"JOURNAL_DB="+dbPath,
	)

	// Add a caucion
	addCmd := exec.Command(bin, "add",
		"--principal", "500000",
		"--tna", "90",
		"--term", "7",
		"--date", "2026-03-01",
		"--notes", "sqlite test",
	)
	addCmd.Env = env
	out, err := addCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("add sqlite: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "Saved caucion") {
		t.Fatalf("expected saved, got: %s", out)
	}

	// List
	listCmd := exec.Command(bin, "list")
	listCmd.Env = env
	out, err = listCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("list sqlite: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "sqlite test") {
		t.Fatalf("expected caucion in list, got: %s", out)
	}
}
