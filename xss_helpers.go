package libinjection

import (
	"strings"
)

// maxNormalizedTokenLen is the stack buffer size for uppercased, null-stripped
// tag and attribute names. Must exceed the longest blacklisted name
// (currently "ON" + "WEBKITCURRENTPLAYBACKTARGETISWIRELESSCHANGED" = 48).
const maxNormalizedTokenLen = 64

func isH5White(ch byte) bool {
	return ch == '\n' || ch == '\t' || ch == '\v' || ch == '\f' || ch == '\r' || ch == ' '
}

// asciiEqualFold compares two equal-length ASCII strings case-insensitively
// without allocating.
func asciiEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 0x20
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 0x20
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// upperRemoveNulls normalizes s into buf: uppercases ASCII and removes null bytes.
// Returns the number of bytes written and whether the input was truncated.
// Truncation occurs when the number of non-null bytes in s exceeds len(buf)
// (maxNormalizedTokenLen = 64). Any bytes beyond that limit are silently dropped.
// Since all blacklisted tag/attribute names are at most 48 bytes, truncation does
// not affect detection accuracy for any name currently in the blacklist; however,
// callers should check the truncated return value to avoid false negatives if the
// blacklist grows beyond 64 bytes in the future, or to skip further checks when
// a clearly over-length token cannot possibly match.
func upperRemoveNulls(buf []byte, s string) (n int, truncated bool) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0 {
			continue
		}
		if n == len(buf) {
			truncated = true
			break
		}
		if c >= 'a' && c <= 'z' {
			c -= 0x20
		}
		buf[n] = c
		n++
	}
	return n, truncated
}

//nolint:gocyclo // complexity 12, reduction tracked in #122
func isBlackTag(s string) bool {
	if len(s) < 3 {
		return false
	}

	var buf [maxNormalizedTokenLen]byte
	n, truncated := upperRemoveNulls(buf[:], s)
	if truncated {
		// Input is longer than any blacklisted tag name; cannot match.
		return false
	}
	normalized := buf[:n]

	for i := 0; i < len(blackTags); i++ {
		if string(normalized) == blackTags[i] {
			return true
		}
	}

	// anything SVG or XSL(t) related (prefix match on first 3 chars)
	if n >= 3 && ((normalized[0] == 'S' && normalized[1] == 'V' && normalized[2] == 'G') ||
		(normalized[0] == 'X' && normalized[1] == 'S' && normalized[2] == 'L')) {
		return true
	}

	return false
}

//nolint:gocyclo // complexity 12, reduction tracked in #122
func isBlackAttr(s string) int {
	var buf [maxNormalizedTokenLen]byte
	n, truncated := upperRemoveNulls(buf[:], s)
	if truncated {
		// Input is longer than any blacklisted attribute name; cannot match.
		return attributeTypeNone
	}

	if n < 2 {
		return attributeTypeNone
	}
	normalized := buf[:n]

	if n >= 5 {
		if string(normalized) == "XMLNS" || string(normalized) == "XLINK" {
			// got xmlns or xlink tags
			return attributeTypeBlack
		}
		// JavaScript on.* event handlers — O(1) map lookup replaces O(432) scan.
		// Go elides the string([]byte) allocation when used solely as a map key.
		if buf[0] == 'O' && buf[1] == 'N' {
			if typ, ok := blackEventsMap[string(buf[2:n])]; ok {
				return typ
			}
		}
	}

	// O(1) map lookup replaces O(20) scan.
	if typ, ok := blacksMap[string(normalized)]; ok {
		return typ
	}
	return attributeTypeNone
}

// urlStrippedEntities are the only named entities that htmlDecodeByteAt
// decodes: they produce characters that browsers strip from URLs (WHATWG URL
// parsing), so leaving them encoded would let "ja&Tab;vascript:" bypass
// isBlackURL. They match case-sensitively and only with the trailing ';', as
// in the WHATWG entity table: "&tab;" and "&Tab" are not decoded by browsers.
var urlStrippedEntities = []struct {
	name string
	val  int
}{
	{"&Tab;", '\t'},
	{"&NewLine;", '\n'},
}

// decodeStrippedEntity returns the character and the number of bytes consumed
// when s starts with one of urlStrippedEntities, or (0, 0) otherwise.
func decodeStrippedEntity(s string) (int, int) {
	for _, e := range urlStrippedEntities {
		if strings.HasPrefix(s, e.name) {
			return e.val, len(e.name)
		}
	}
	return 0, 0
}

//nolint:gocyclo // complexity 21, reduction tracked in #125
func htmlDecodeByteAt(s string) (int, int) {
	length := len(s)
	val := 0

	if length == 0 {
		return byteEOF, 0
	}

	if s[0] != '&' || length < 2 {
		return int(s[0]), 1
	}

	if s[1] != '#' || len(s) < 3 {
		if ch, consumed := decodeStrippedEntity(s); consumed > 0 {
			return ch, consumed
		}
		// normally this would be for named entities
		// but for this case we don't actually care
		return '&', 1
	}

	if s[2] == 'x' || s[2] == 'X' {
		if len(s) < 4 {
			return '&', 1
		}
		ch := int(s[3])
		ch = gsHexDecodeMap[ch]
		if ch == 256 {
			// degenerate case '&#[?]'
			return '&', 1
		}
		val = ch
		i := 4

		for i < length {
			ch = int(s[i])
			if ch == ';' {
				return val, i + 1
			}
			ch = gsHexDecodeMap[ch]
			if ch == 256 {
				return val, i
			}
			val = val*16 + ch
			if val > 0x1000FF {
				return '&', 1
			}
			i++
		}
		return val, i
	}
	i := 2
	ch := int(s[i])
	if ch < '0' || ch > '9' {
		return '&', 1
	}
	val = ch - '0'
	i++
	for i < length {
		ch = int(s[i])
		if ch == ';' {
			return val, i + 1
		}
		if ch < '0' || ch > '9' {
			return val, i
		}
		val = val*10 + (ch - '0')
		if val > 0x1000FF {
			return '&', 1
		}
		i++
	}
	return val, i
}

// Does an HTML encoded  binary string (const char*, length) start with
// a all uppercase c-string (null terminated), case insensitive!
//
// also ignore any embedded nulls, TAB, LF and CR in the HTML string, since
// browsers strip TAB, LF and CR anywhere in a URL! This helper is only used
// for URL scheme prefixes (see isBlackURL).
//
//nolint:gocyclo // complexity 10, reduction tracked in #122
func htmlEncodeStartsWith(a, b string) bool {
	var (
		first  = true
		pos    = 0
		length = len(b)
		ai     = 0
	)

	for length > 0 {
		if ai >= len(a) {
			// already matched the full prefix
			return true
		}
		cb, consumed := htmlDecodeByteAt(b[pos:])
		pos += consumed
		length -= consumed

		if first && cb <= 32 {
			// ignore all leading whitespace and control characters
			continue
		}
		first = false

		if cb == 0 || cb == 9 || cb == 10 || cb == 13 {
			// ignore nulls; browsers strip TAB, LF and CR anywhere in a URL
			continue
		}
		if cb > 255 {
			// A reference above 0xFF is no ASCII letter. C truncates it to a
			// char (&#330; -> 'J'); do not do that here.
			return false
		}
		if cb >= 'a' && cb <= 'z' {
			cb -= 0x20
		}

		// cb is in 0..255 here; the mask only makes the conversion explicit
		if byte(cb&0xFF) != a[ai] {
			return false
		}
		ai++
	}

	return ai >= len(a)
}

func isBlackURL(s string) bool {
	urls := []string{
		"DATA",        // data url
		"VIEW-SOURCE", // view source url
		"VBSCRIPT",    // obsolete but interesting signal
		"JAVA",        // covers JAVA, JAVASCRIPT, + colon
	}

	//  HEY: this is a signed character.
	//  We are intentionally skipping high-bit characters too
	//  since they are not ASCII, and Opera sometimes uses UTF-8 whitespace.
	//
	//  Also in EUC-JP some of the high bytes are just ignored.
	str := strings.TrimLeftFunc(s, func(r rune) bool {
		return r <= 32 || r >= 127
	})

	for _, url := range urls {
		if htmlEncodeStartsWith(url, str) {
			return true
		}
	}
	return false
}
