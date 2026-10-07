package libinjection

import (
	"strings"
	"testing"
	"time"
)

// TestParseStringLinear guards against GHSA-974q-67pr-5pc2: parseStringCore and
// isBackslashEscaped used to rescan an ever-growing prefix of the input on every
// iteration, making IsSQLi O(n^2) in the length of a quote-heavy string literal.
// A ~1 MB payload took ~100s before the fix and runs in milliseconds after it.
func TestParseStringLinear(t *testing.T) {
	const budget = 10 * time.Second
	n := 500_000
	payloads := map[string]string{
		"backslash-escaped quote": "'" + strings.Repeat(`\'`, n),
		"doubled quote":           "'" + strings.Repeat(`''`, n),
	}
	for name, payload := range payloads {
		t.Run(name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				IsSQLi(payload)
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(budget):
				t.Fatalf("IsSQLi on a %d-byte %s payload did not finish within %s; "+
					"parseStringCore is scaling super-linearly", len(payload), name, budget)
			}
		})
	}
}
