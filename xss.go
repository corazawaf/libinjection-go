package libinjection

func isXSS(input string, flags int) bool {
	var (
		h5   = new(h5State)
		attr = attributeTypeNone
	)

	h5.init(input, flags)
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
			// Use direct byte search instead of strings.IndexByte
			hasBacktick := false
			for i := 0; i < h5.tokenLen; i++ {
				if h5.tokenStart[i] == '`' {
					hasBacktick = true
					break
				}
			}
			if hasBacktick {
				return true
			}

			// IE conditional comment
			if h5.tokenLen > 3 {
				if h5.tokenStart[0] == '[' &&
					(h5.tokenStart[1] == 'I' || h5.tokenStart[1] == 'i') &&
					(h5.tokenStart[2] == 'F' || h5.tokenStart[2] == 'f') {
					return true
				}

				if (h5.tokenStart[1] == 'X' || h5.tokenStart[1] == 'x') &&
					(h5.tokenStart[2] == 'M' || h5.tokenStart[2] == 'm') &&
					(h5.tokenStart[3] == 'L' || h5.tokenStart[3] == 'l') {
					return true
				}
			}

			if h5.tokenLen > 5 {
				// Check for IMPORT or ENTITY without string allocations
				if h5.tokenLen >= 6 {
					// Check for IMPORT
					if (h5.tokenStart[0] == 'I' || h5.tokenStart[0] == 'i') &&
						(h5.tokenStart[1] == 'M' || h5.tokenStart[1] == 'm') &&
						(h5.tokenStart[2] == 'P' || h5.tokenStart[2] == 'p') &&
						(h5.tokenStart[3] == 'O' || h5.tokenStart[3] == 'o') &&
						(h5.tokenStart[4] == 'R' || h5.tokenStart[4] == 'r') &&
						(h5.tokenStart[5] == 'T' || h5.tokenStart[5] == 't') {
						return true
					}
					// Check for ENTITY
					if (h5.tokenStart[0] == 'E' || h5.tokenStart[0] == 'e') &&
						(h5.tokenStart[1] == 'N' || h5.tokenStart[1] == 'n') &&
						(h5.tokenStart[2] == 'T' || h5.tokenStart[2] == 't') &&
						(h5.tokenStart[3] == 'I' || h5.tokenStart[3] == 'i') &&
						(h5.tokenStart[4] == 'T' || h5.tokenStart[4] == 't') &&
						(h5.tokenStart[5] == 'Y' || h5.tokenStart[5] == 'y') {
						return true
					}
				}
			}
		}
	}

	return false
}

// IsXSS returns true if the input string contains XSS
func IsXSS(input string) bool {
	if isXSS(input, html5FlagsDataState) ||
		isXSS(input, html5FlagsValueNoQuote) ||
		isXSS(input, html5FlagsValueSingleQuote) ||
		isXSS(input, html5FlagsValueDoubleQuote) ||
		isXSS(input, html5FlagsValueBackQuote) {
		return true
	}

	return false
}
