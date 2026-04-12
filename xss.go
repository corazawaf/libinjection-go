package libinjection

import (
	"strings"
	"sync"
)

var h5StatePool = sync.Pool{New: func() any { return new(h5State) }}

func isXSS(input string, flags int) bool {
	h5 := h5StatePool.Get().(*h5State)
	defer h5StatePool.Put(h5)
	h5.init(input, flags) // full reset then re-init
	return runXSS(h5)
}

// runXSS contains the detection loop; split out so isXSS can defer the pool
// Put before the loop runs, keeping the fast-return paths clean.
func runXSS(h5 *h5State) bool {
	attr := attributeTypeNone
	for h5.next() {
		if h5.tokenType != html5TypeAttrValue {
			attr = attributeTypeNone
		}

		switch h5.tokenType {

		case html5TypeDocType:
			return true
		case html5TypeTagNameOpen:
			if isBlackTag(h5.tokenStart[:h5.tokenLen]) {
				return true
			}
		case html5TypeAttrName:
			attr = isBlackAttr(h5.tokenStart[:h5.tokenLen])
		case html5TypeAttrValue:
			// IE6,7,8 parsing works a bit differently so
			// a whole <script> or other black tag might be hiding
			// inside an attribute value under HTML 5 parsing
			// See http://html5sec.org/#102
			// to avoid doing a full reparse of the value, just
			// look for "<".  This probably need adjusting to
			// handle escaped characters
			switch attr {
			case attributeTypeNone:
				break
			case attributeTypeBlack:
				return true
			case attributeTypeAttrURL:
				if isBlackURL(h5.tokenStart[:h5.tokenLen]) {
					return true
				}
			case attributeTypeStyle:
				return true
			case attributeTypeAttrIndirect:
				// an attribute name is specified in a _value_
				if isBlackAttr(h5.tokenStart[:h5.tokenLen]) == attributeTypeBlack {
					return true
				}
			}
			attr = attributeTypeNone
		case html5TypeTagComment:
			// IE uses a "`" as a tag ending byte
			if strings.IndexByte(h5.tokenStart[:h5.tokenLen], '`') != -1 {
				return true
			}

			// IE conditional comment
			if h5.tokenLen > 3 {
				if h5.tokenStart[0] == '[' &&
					(h5.tokenStart[1] == 'I' || h5.tokenStart[1] == 'i') &&
					(h5.tokenStart[2] == 'F' || h5.tokenStart[2] == 'f') {
					return true
				}

				if (h5.tokenStart[0] == 'X' || h5.tokenStart[0] == 'x') &&
					(h5.tokenStart[1] == 'M' || h5.tokenStart[1] == 'm') &&
					(h5.tokenStart[2] == 'L' || h5.tokenStart[2] == 'l') {
					return true
				}
			}

			if h5.tokenLen > 5 {
				var buf [6]byte
				n, _ := upperRemoveNulls(buf[:], h5.tokenStart[:6])

				// IE <?import pseudo-tag or XML Entity definition
				if n == 6 && (string(buf[:6]) == "IMPORT" || string(buf[:6]) == "ENTITY") {
					return true
				}
			}
		}
	}

	return false
}

// IsXSS returns true if the input string contains XSS.
//
// Five HTML5 parse contexts are tried. The DataState context requires '<' to
// produce any tag tokens, so it is skipped when '<' is absent — saving one
// full state-machine pass for the common case of clean input. The four
// attribute-value contexts can detect injection without '<' (e.g. onerror=...)
// and always run.
func IsXSS(input string) bool {
	// DataState requires '<'; skip it when absent to save one pass.
	if strings.IndexByte(input, '<') != -1 {
		if isXSS(input, html5FlagsDataState) {
			return true
		}
	}
	return isXSS(input, html5FlagsValueNoQuote) ||
		isXSS(input, html5FlagsValueSingleQuote) ||
		isXSS(input, html5FlagsValueDoubleQuote) ||
		isXSS(input, html5FlagsValueBackQuote)
}
