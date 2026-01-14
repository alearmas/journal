package service_test

import (
	"alearmas/tradingJournal/internal/service"
	"testing"
)

func FuzzParseDecimal(f *testing.F) {
	// Seeds: casos interesantes conocidos
	f.Add("1000000.00")
	f.Add("0")
	f.Add("-1")
	f.Add("85.5")
	f.Add("1e1000")
	f.Add("NaN")
	f.Add("")
	f.Add("abc")

	f.Fuzz(func(t *testing.T, input string) {
		_, _ = service.ParseDecimal(input)
		// No assertions.
		// The test FAILS if this panics.
	})
}
