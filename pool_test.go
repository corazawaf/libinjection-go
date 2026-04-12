package libinjection

import (
	"sync"
	"testing"
)

// TestSQLiPoolReuse verifies that pooled sqliState objects do not leak state
// between consecutive calls. Alternating attack / clean inputs must each
// return the correct result regardless of what the previous call did.
func TestSQLiPoolReuse(t *testing.T) {
	cases := []struct {
		input  string
		isSQLi bool
	}{
		{`1 UNION SELECT username, password FROM users--`, true},
		{`hello world`, false},
		{`1' AND 1=1--`, true},
		{`user@example.com`, false},
		{`1'; DROP TABLE users--`, true},
		{`2024-01-15`, false},
		{`1/**/UNION/**/SELECT/**/1,2,3--`, true},
		{`The quick brown fox jumps over the lazy dog`, false},
		// Repeat the same attack twice: pool returns same object second time.
		{`1 UNION SELECT username, password FROM users--`, true},
		{`1 UNION SELECT username, password FROM users--`, true},
		// Repeat clean twice.
		{`hello world`, false},
		{`hello world`, false},
	}

	for _, tc := range cases {
		got, _ := IsSQLi(tc.input)
		if got != tc.isSQLi {
			t.Errorf("IsSQLi(%q) = %v, want %v", tc.input, got, tc.isSQLi)
		}
	}
}

// TestXSSPoolReuse verifies that pooled h5State objects do not leak state
// between consecutive IsXSS calls.
func TestXSSPoolReuse(t *testing.T) {
	cases := []struct {
		input string
		isXSS bool
	}{
		{`<script>alert(1)</script>`, true},
		{`<p>Hello world</p>`, false},
		{`<img src=x onerror=alert(1)>`, true},
		{`normal text without any html`, false},
		{`<svg onload=alert(1)>`, true},
		{`john.doe@example.com`, false},
		// Repeat the same attack twice.
		{`<script>alert(1)</script>`, true},
		{`<script>alert(1)</script>`, true},
		// Repeat clean twice.
		{`<p>Hello world</p>`, false},
		{`<p>Hello world</p>`, false},
	}

	for _, tc := range cases {
		got := IsXSS(tc.input)
		if got != tc.isXSS {
			t.Errorf("IsXSS(%q) = %v, want %v", tc.input, got, tc.isXSS)
		}
	}
}

// TestXSSDataStatePrefilter documents the DataState '<' prefilter behaviour:
//   - Inputs that contain XSS in an attribute-value context (no '<') are still
//     detected by the four attribute-value parse contexts.
//   - The DataState pass is simply skipped when '<' is absent; detection
//     correctness is not compromised.
func TestXSSDataStatePrefilter(t *testing.T) {
	// Detected via attribute-value contexts (no '<' required).
	noAngleAttacks := []string{
		`onerror=alert(1)`,
		`onerror=alert(1)>`,
		`x onerror=alert(1);>`,
		`x' onerror=alert(1);>`,
		`x" onerror=alert(1);>`,
		`onload=alert(1)`,
		`onclick=alert(1)`,
	}
	for _, input := range noAngleAttacks {
		if !IsXSS(input) {
			t.Errorf("IsXSS(%q) = false, want true (attribute-value context)", input)
		}
	}

	// Clean inputs without '<' must not trigger false positives.
	noAngleClean := []string{
		`hello world`,
		`john.doe@example.com`,
		`onY29va2llcw==`,
		`myvar=onfoobar==`,
		`2024-01-15`,
	}
	for _, input := range noAngleClean {
		if IsXSS(input) {
			t.Errorf("IsXSS(%q) = true, want false (clean input, no '<')", input)
		}
	}
}

// TestPoolConcurrency verifies that pooled state objects are safe under
// concurrent access. WAF deployments call IsSQLi and IsXSS from many
// goroutines simultaneously; this test exercises that path with both attack
// and clean inputs to catch any state leakage between goroutines.
func TestPoolConcurrency(t *testing.T) {
	t.Parallel()

	const goroutines = 50
	const iterations = 200

	t.Run("SQLi", func(t *testing.T) {
		t.Parallel()
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()
				for range iterations {
					if got, _ := IsSQLi(`1 UNION SELECT 1,2--`); !got {
						t.Error("IsSQLi: expected true for attack input")
					}
					if got, _ := IsSQLi(`hello world`); got {
						t.Error("IsSQLi: expected false for clean input")
					}
				}
			}()
		}
		wg.Wait()
	})

	t.Run("XSS", func(t *testing.T) {
		t.Parallel()
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()
				for range iterations {
					if !IsXSS(`<script>alert(1)</script>`) {
						t.Error("IsXSS: expected true for attack input")
					}
					if IsXSS(`hello world`) {
						t.Error("IsXSS: expected false for clean input")
					}
				}
			}()
		}
		wg.Wait()
	})
}
