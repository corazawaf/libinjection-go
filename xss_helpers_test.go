package libinjection

import (
	"strings"
	"testing"
)

func TestAsciiEqualFold(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{name: "equal length, exact match", a: "doctype", b: "doctype", want: true},
		{name: "equal length, a uppercase", a: "DOCTYPE", b: "doctype", want: true},
		{name: "equal length, b uppercase", a: "doctype", b: "DOCTYPE", want: true},
		{name: "equal length, mismatch", a: "data", b: "date", want: false},
		{name: "different length", a: "data", b: "dat", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := asciiEqualFold(tt.a, tt.b); got != tt.want {
				t.Errorf("asciiEqualFold(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestHtmlDecodeByteAt(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantVal      int
		wantConsumed int
	}{
		{name: "empty string", input: "", wantVal: byteEOF, wantConsumed: 0},
		{name: "regular char", input: "a", wantVal: 'a', wantConsumed: 1},
		{name: "ampersand only", input: "&", wantVal: '&', wantConsumed: 1},
		{name: "ampersand with non-hash", input: "&a", wantVal: '&', wantConsumed: 1},
		{name: "&#  too short", input: "&#", wantVal: '&', wantConsumed: 1},
		{name: "&#non-digit", input: "&#a", wantVal: '&', wantConsumed: 1},
		{name: "decimal no terminator", input: "&#5", wantVal: 5, wantConsumed: 3},
		{name: "decimal semicolon", input: "&#5;", wantVal: 5, wantConsumed: 4},
		{name: "decimal non-digit terminator", input: "&#5a", wantVal: 5, wantConsumed: 3},
		{name: "decimal overflow", input: "&#9999999", wantVal: '&', wantConsumed: 1},
		{name: "hex &#x too short", input: "&#x", wantVal: '&', wantConsumed: 1},
		{name: "hex uppercase X too short", input: "&#X", wantVal: '&', wantConsumed: 1},
		{name: "hex invalid char", input: "&#xG", wantVal: '&', wantConsumed: 1},
		{name: "hex no terminator", input: "&#x5", wantVal: 5, wantConsumed: 4},
		{name: "hex semicolon", input: "&#x5;", wantVal: 5, wantConsumed: 5},
		{name: "hex uppercase X valid", input: "&#X5", wantVal: 5, wantConsumed: 4},
		{name: "hex non-hex terminator", input: "&#x5G", wantVal: 5, wantConsumed: 4},
		{name: "hex overflow", input: "&#x1000FF5", wantVal: '&', wantConsumed: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotConsumed := htmlDecodeByteAt(tt.input)
			if gotVal != tt.wantVal || gotConsumed != tt.wantConsumed {
				t.Errorf("htmlDecodeByteAt(%q) = (%d, %d), want (%d, %d)",
					tt.input, gotVal, gotConsumed, tt.wantVal, tt.wantConsumed)
			}
		})
	}
}

func TestIsBlackAttr(t *testing.T) {
	tests := []struct {
		name string
		attr string
		want int
	}{
		{
			name: "Test with black attribute",
			attr: "xmlns",
			want: attributeTypeBlack,
		},
		{
			name: "Test with non-black attribute",
			attr: "class",
			want: attributeTypeNone,
		},
		{
			name: "Test with JavaScript event handler",
			attr: "onclick",
			want: attributeTypeBlack,
		},
		{
			name: "Test with short attribute",
			attr: "a",
			want: attributeTypeNone,
		},
		{
			name: "Test with long null attribute that will be stripped",
			attr: "a\x00\x00\x00\x00\x00",
			want: attributeTypeNone,
		},
		{
			name: "over-length attribute cannot match",
			attr: "onclick" + strings.Repeat("x", maxNormalizedTokenLen), // 7+64 = 71 bytes > maxNormalizedTokenLen
			want: attributeTypeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBlackAttr(tt.attr); got != tt.want {
				t.Errorf("isBlackAttr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHtmlEncodeStartsWith(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		input  string
		want   bool
	}{
		{name: "exact match", prefix: "DATA", input: "data:", want: true},
		{name: "prefix match with trailing content", prefix: "JAVA", input: "javascript:alert(1)", want: true},
		{name: "no match", prefix: "DATA", input: "https://example.com", want: false},
		{name: "pattern in middle should not match", prefix: "DATA", input: "https://github.com/Simbiat/database", want: false},
		{name: "pattern at end should not match", prefix: "DATA", input: "nodata", want: false},
		{name: "leading whitespace skipped", prefix: "DATA", input: "\tdata:", want: true},
		{name: "embedded null ignored", prefix: "DATA", input: "d\x00ata:", want: true},
		{name: "embedded LF ignored", prefix: "DATA", input: "d\nata:", want: true},
		{name: "input shorter than prefix", prefix: "DATA", input: "dat", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := htmlEncodeStartsWith(tt.prefix, tt.input); got != tt.want {
				t.Errorf("htmlEncodeStartsWith(%q, %q) = %v, want %v", tt.prefix, tt.input, got, tt.want)
			}
		})
	}
}

func TestIsBlackURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "data URL", url: "data:text/html,<script>alert(1)</script>", want: true},
		{name: "javascript URL", url: "javascript:alert(1)", want: true},
		{name: "vbscript URL", url: "vbscript:msgbox", want: true},
		{name: "https URL", url: "https://example.com", want: false},
		{name: "URL containing data in path", url: "https://github.com/Simbiat/database", want: false},
		{name: "URL containing java in path", url: "https://example.com/javascript-tutorials", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBlackURL(tt.url); got != tt.want {
				t.Errorf("isBlackURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsBlackTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{name: "script tag", tag: "SCRIPT", want: true},
		{name: "svg exact", tag: "svg", want: true},
		{name: "SVG uppercase", tag: "SVG", want: true},
		{name: "svg prefixed tag", tag: "svganimate", want: true},
		{name: "svg namespaced", tag: "svg:rect", want: true},
		{name: "xsl exact", tag: "xsl", want: true},
		{name: "xsl prefixed tag", tag: "xsl:template", want: true},
		{name: "div tag", tag: "div", want: false},
		{name: "span tag", tag: "span", want: false},
		{name: "too short", tag: "sv", want: false},
		{name: "over-length tag cannot match", tag: "script" + strings.Repeat("x", maxNormalizedTokenLen), want: false},          // 6+64 = 70 bytes > maxNormalizedTokenLen
		{name: "over-length SVG prefix cannot match via prefix rule", tag: "svg" + strings.Repeat("x", maxNormalizedTokenLen), want: false}, // 3+64 = 67 bytes > maxNormalizedTokenLen
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBlackTag(tt.tag); got != tt.want {
				t.Errorf("isBlackTag(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}
