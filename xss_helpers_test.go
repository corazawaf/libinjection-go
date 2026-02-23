package libinjection

import (
	"strings"
	"testing"
)

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
