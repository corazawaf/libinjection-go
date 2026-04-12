package libinjection

import (
	"testing"
)

// sqliPayloads are representative inputs for the @detectSQLi operator — a mix
// of attack payloads and clean traffic that reflects what a WAF processes in
// production (CRS rules 942100 and similar).
var sqliPayloads = []string{
	// Classic UNION-based
	`1 UNION SELECT username, password FROM users--`,
	// Boolean blind
	`1' AND 1=1--`,
	// Stacked queries
	`1'; DROP TABLE users--`,
	// Comment-based evasion
	`1/**/UNION/**/SELECT/**/1,2,3--`,
	// Time-based blind
	`1' AND SLEEP(5)--`,
	// Double quote context
	`" OR "1"="1`,
	// Clean input (false positive check, dominant traffic)
	`hello world`,
	`SELECT * FROM products WHERE id = 42`,
	`user@example.com`,
	`2024-01-15`,
	// Encoded/obfuscated
	`1%27+AND+1%3D1--`,
	// Long benign input
	`The quick brown fox jumps over the lazy dog near the riverbank at sunset`,
}

// xssPayloads are representative inputs for the @detectXSS operator — a mix
// of attack payloads and clean traffic (CRS rules 941100 and similar).
var xssPayloads = []string{
	// Script tag
	`<script>alert(1)</script>`,
	// Event handler
	`<img src=x onerror=alert(1)>`,
	// SVG XSS
	`<svg onload=alert(1)>`,
	// JavaScript href
	`<a href="javascript:alert(1)">click</a>`,
	// Encoded
	`<script>alert&#40;1&#41;</script>`,
	// Clean HTML (dominant traffic)
	`<p>Hello world</p>`,
	`<div class="container"><h1>Title</h1></div>`,
	`normal text without any html`,
	// Data URI
	`<img src="data:text/html,<script>alert(1)</script>">`,
}

func BenchmarkIsSQLi_CRS(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range sqliPayloads {
			IsSQLi(input)
		}
	}
}

func BenchmarkIsXSS_CRS(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range xssPayloads {
			IsXSS(input)
		}
	}
}

// Per-input benchmarks for profiling hot paths
func BenchmarkIsSQLi_Clean(b *testing.B) {
	input := "hello world"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsSQLi(input)
	}
}

func BenchmarkIsSQLi_Union(b *testing.B) {
	input := `1 UNION SELECT username, password FROM users--`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsSQLi(input)
	}
}

func BenchmarkIsSQLi_Boolean(b *testing.B) {
	input := `1' AND 1=1--`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsSQLi(input)
	}
}

func BenchmarkIsSQLi_CommentEvasion(b *testing.B) {
	input := `1/**/UNION/**/SELECT/**/1,2,3--`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsSQLi(input)
	}
}

func BenchmarkIsXSS_Clean(b *testing.B) {
	input := "<p>Hello world</p>"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsXSS(input)
	}
}

func BenchmarkIsXSS_ScriptTag(b *testing.B) {
	input := "<script>alert(1)</script>"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsXSS(input)
	}
}

func BenchmarkIsXSS_EventHandler(b *testing.B) {
	input := `<img src=x onerror=alert(1)>`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsXSS(input)
	}
}

// BenchmarkIsXSS_NoAngle shows the DataState prefilter: input without '<'
// skips the DataState pass but still runs the 4 attribute-value contexts.
func BenchmarkIsXSS_NoAngle(b *testing.B) {
	input := `onerror=alert(1)` // attribute-injection context, no '<'
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsXSS(input)
	}
}

// BenchmarkIsXSS_PlainText measures the dominant WAF case: a clean param
// with no HTML markup at all.
func BenchmarkIsXSS_PlainText(b *testing.B) {
	input := `john.doe@example.com`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsXSS(input)
	}
}
