package libinjection

import "strings"

func isH5White(ch byte) bool {
	if ch == '\n' || ch == '\t' || ch == '\v' || ch == '\f' || ch == '\r' {
		return true
	} else {
		return false
	}
}

func isBlackTag(s string) bool {
	if len(s) < 3 {
		return false
	}

	for i := 0; i < len(blackTags); i++ {
		if strings.ToUpper(strings.ReplaceAll(s, "\x00", "")) == blackTags[i] {
			return true
		}
	}

	// anything SVG related
	if strings.ToUpper(s) == "SVG" {
		return true
	}

	// anything XSL(t) related
	if strings.ToUpper(s) == "XSL" {
		return true
	}

	return false
}

func isBlackAttr(s string) int {
	length := len(s)
	if length < 2 {
		return attributeTypeNone
	}

	if length >= 5 {
		// javascript on.*
		if strings.ToUpper(s[:2]) == "ON" {
			// got javascript on- attribute name
			return attributeTypeBlack
		}

		if strings.ToUpper(strings.ReplaceAll(s, "\x00", "")) == "XMLNS" ||
			strings.ToUpper(strings.ReplaceAll(s, "\x00", "")) == "XLINK" {
			// got xmlns or xlink tags
			return attributeTypeBlack
		}
	}

	for _, black := range blacks {
		if strings.ToUpper(strings.ReplaceAll(s, "\x00", "")) == black.name {
			// got banner attribute name
			return black.attributeType
		}
	}
	return attributeTypeNone
}

// Does an HTML encoded  binary string (const char*, length) start with
// a all uppercase c-string (null terminated), case insensitive!
//
// also ignore any embedded nulls in the HTML string!
// todo: implement
func htmlEncodeStartsWith(a, b string) bool {
	return false
}

func isBlackUrl(s string) bool {
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
	space := strings.TrimLeftFunc(s, func(r rune) bool {
		return r <= 32 || r >= 127
	})
	str := s[len(space):]

	for _, url := range urls {
		if htmlEncodeStartsWith(url, str) {
			return true
		}
	}
	return false
}
