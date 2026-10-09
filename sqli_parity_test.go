package libinjection

import "testing"

// TestUpstreamParity pins inputs whose verdict and fingerprint diverged from
// upstream C libinjection. Expected values were produced by the C
// implementation (libinjection/libinjection master, x86-64, signed char).
func TestUpstreamParity(t *testing.T) {
	cases := []struct {
		name, input, fingerprint string
		detected                 bool
	}{
		// notWhitelist: a 2-token "1c" whose comment starts with '/' is SQLi;
		// the check used to be inverted, and the fallback that happened to
		// catch "1/*" assumes the number sits at offset 0.
		{"1c block comment, leading tab", "\t1/*", "1c", true},
		{"1c block comment, hex number, leading space", " 0x1/*", "1c", true},
		{"1c block comment", "1/*", "1c", true},
		{"1c dash comment", "1--", "1c", true},
		// Numeric hash comments intentionally depart from C's suppression;
		// their hardened verdict is covered by TestIsSQLiSecurityRegressions.
		// parseQStringCore: a q-string delimiter byte >= 0x80 must fall back to
		// parseWord, as it does in C where char is signed.
		{"q-string delimiter above 0x7f", "q'\xe9' union /*!50000select*/ 1", "X", true},
		// parseStringCore: the closing quote is judged at its real position,
		// not at the first occurrence of the same suffix text.
		{"backslash-escaped quote at real position", `-(top<>thendrop1.5\'--null'--`, "sc", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			detected, fp := IsSQLi(c.input)
			if detected != c.detected || fp != c.fingerprint {
				t.Errorf("IsSQLi(%q) = (%v, %q), want (%v, %q)", c.input, detected, fp, c.detected, c.fingerprint)
			}
		})
	}
}
