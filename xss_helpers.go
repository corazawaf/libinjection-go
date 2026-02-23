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
// Returns the number of bytes written.
func upperRemoveNulls(buf []byte, s string) int {
	n := 0
	for i := 0; i < len(s) && n < len(buf); i++ {
		c := s[i]
		if c == 0 {
			continue
		}
		if c >= 'a' && c <= 'z' {
			c -= 0x20
		}
		buf[n] = c
		n++
	}
	return n
}

func isBlackTag(s string) bool {
	if len(s) < 3 {
		return false
	}

	var buf [maxNormalizedTokenLen]byte
	n := upperRemoveNulls(buf[:], s)
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

func isBlackAttr(s string) int {
	var buf [maxNormalizedTokenLen]byte
	n := upperRemoveNulls(buf[:], s)

	if n < 2 {
		return attributeTypeNone
	}
	normalized := buf[:n]

	if n >= 5 {
		if string(normalized) == "XMLNS" || string(normalized) == "XLINK" {
			// got xmlns or xlink tags
			return attributeTypeBlack
		}
		// JavaScript on.* event handlers
		if buf[0] == 'O' && buf[1] == 'N' {
			eventName := buf[2:n]
			// got javascript on- attribute name
			for _, event := range blackEvents {
				if string(eventName) == event.name {
					return event.attributeType
				}
			}
		}
	}

	for _, black := range blacks {
		if string(normalized) == black.name {
			// got banner attribute name
			return black.attributeType
		}
	}
	return attributeTypeNone
}

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
// also ignore any embedded nulls in the HTML string!
func htmlEncodeStartsWith(a, b string) bool {
	var (
		first  = true
		pos    = 0
		length = len(b)
		ai     = 0
	)

	for length > 0 {
		cb, consumed := htmlDecodeByteAt(b[pos:])
		pos += consumed
		length -= consumed

		if first && cb <= 32 {
			// ignore all leading whitespace and control characters
			continue
		}
		first = false

		if cb == 0 || cb == 10 {
			// always ignore null characters in user input
			// always ignore vertical tab characters in user input
			continue
		}
		if cb >= 'a' && cb <= 'z' {
			cb -= 0x20
		}
		// Mask to 8 bits to match C's implicit char truncation behavior.
		ch := byte(cb & 0xFF)

		if ai >= len(a) {
			// already matched the full prefix
			return true
		}
		if ch != a[ai] {
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
