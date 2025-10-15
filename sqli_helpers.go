package libinjection

import (
	"unsafe"
)

func flag2Delimiter(flag int) byte {
	switch {
	case (flag & sqliFlagQuoteSingle) != 0:
		return byteSingle
	case (flag & sqliFlagQuoteDouble) != 0:
		return byteDouble
	default:
		return byteNull
	}
}

// OK! "	\"	" one backslash = escaped!
//
//		   "   \\"	" two backslash = not escaped!
//	    "  \\\"	" three backslash = escaped!
func isBackslashEscaped(str string) bool {
	// Check if there's any backslash using idiomatic Go
	hasBackslash := false
	for _, ch := range str {
		if ch == '\\' {
			hasBackslash = true
			break
		}
	}
	if !hasBackslash {
		return false
	}

	// Count trailing backslashes from the end
	count := 0
	for i := len(str) - 1; i >= 0 && str[i] == '\\'; i-- {
		count++
	}

	// if number of backslashes is odd, it is escaped
	return count&1 == 1
}

func isDoubleDelimiterEscaped(str string) bool {
	return len(str) >= 2 && str[0] == str[1]
}

func isByteWhite(ch byte) bool {
	// Check for whitespace characters
	// Includes space, tab, newline, vertical tab, form feed, carriage return,
	// null (for Oracle), and Latin-1 non-breaking space
	switch ch {
	case ' ', '\t', '\n', '\v', '\f', '\r', '\x00', '\xa0':
		return true
	default:
		return false
	}
}

// Find the largest string containing certain characters.
//
// if accept is "ABC", then this function would be similar to
// regexp.match(str, "[ABC]*")
func strLenSpn(s string, length int, accept string) int {
	// Use direct byte search - no allocations
	for i := 0; i < length; i++ {
		found := false
		for j := 0; j < len(accept); j++ {
			if s[i] == accept[j] {
				found = true
				break
			}
		}
		if !found {
			return i
		}
	}

	return length
}

func strLenCSpn(s string, length int, accept []byte) int {
	for i := 0; i < length; i++ {
		if accept[s[i]] == 1 {
			return i
		}
	}

	return length
}

// This detects MySQL comments, comments that
// start with /x! We just ban these now but
// previously we attempted to parse the inside.
//
// For reference:
// the form of /x![anything]x/ or /x!12345[anything]x/
//
// MySQL3 (maybe 4), allowed this:
//
//	/x!0selectx/ 1;
//
// where 0 could be any number
//
// The last version of MySQL 3 was in 2003.
//
// It is unclear if the MySQL 3 syntax was allowed
// in MySQL 4. The last version of MySQL 4 was in 2008.
func isMysqlComment(s string, pos int) bool {
	// so far...
	// s[pos] == '/' && s[pos+1] == '*'
	if pos+2 >= len(s) {
		return false
	}

	if s[pos+2] != '!' {
		return false
	}

	return true
}

func toUpperCmp(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	// Compare with case insensitivity - idiomatic Go approach
	for i, aCh := range a {
		bCh := b[i]
		// Convert lowercase to uppercase for comparison
		if bCh >= 'a' && bCh <= 'z' {
			bCh -= 32
		}
		if byte(aCh) != bCh {
			return false
		}
	}
	return true
}

func isKeyword(key string) byte {
	return searchKeyword(key, sqlKeywords)
}

func searchKeyword(key string, keywords map[string]byte) byte {
	// Try direct lookup first (no allocation if already uppercase)
	if val, ok := keywords[key]; ok {
		return val
	}

	// Only convert to uppercase if needed
	hasLower := false
	for i := 0; i < len(key); i++ {
		if key[i] >= 'a' && key[i] <= 'z' {
			hasLower = true
			break
		}
	}

	if !hasLower {
		return byteNull
	}

	// Convert to uppercase
	buf := make([]byte, len(key))
	for i := 0; i < len(key); i++ {
		if key[i] >= 'a' && key[i] <= 'z' {
			buf[i] = key[i] - 32
		} else {
			buf[i] = key[i]
		}
	}

	// Use unsafe conversion to avoid allocation for map lookup
	// This is safe because we're only using it for the map lookup
	upperKey := *(*string)(unsafe.Pointer(&buf))
	if val, ok := keywords[upperKey]; ok {
		return val
	}

	return byteNull
}
