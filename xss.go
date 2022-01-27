package libinjection

import "strings"

func isXss(input string, flags int) bool {
	var (
		h5   = new(h5State)
		attr = attributeTypeNone
	)

	h5.init(input, flags)
	for h5.next() {
		if h5.tokenType != html5TypeAttrValue {
			attr = attributeTypeNone
		}

		if h5.tokenType == html5TypeDocType {
			return true
		} else if h5.tokenType == html5TypeTagNameOpen {
			if isBlackTag(h5.tokenStart[:h5.tokenLen]) {
				return true
			}
		} else if h5.tokenType == html5TypeAttrName {
			attr = isBlackAttr(h5.tokenStart[:h5.tokenLen])
		} else if h5.tokenType == html5TypeAttrValue {
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
				if isBlackUrl(h5.tokenStart[:h5.tokenLen]) {
					return true
				}
				break
			case attributeTypeStyle:
				return true
			case attributeTypeAttrIndirect:
				// an attribute name is specified in a _value_
				if isBlackAttr(h5.tokenStart[:h5.tokenLen]) == attributeTypeBlack {
					return true
				}
				break
			}
			attr = attributeTypeNone
		} else if h5.tokenType == html5TypeTagComment {
			// IE uses a "`" as a tag ending byte
			if strings.IndexByte(h5.tokenStart[:h5.tokenLen], '`') != -1 {
				return true
			}

			// IE conditional comment
			if h5.tokenLen > 3 {
				if h5.tokenStart[0] == '[' &&
					strings.ToUpper(h5.tokenStart[1:3]) == "IF" {
					return true
				}

				if strings.ToUpper(h5.tokenStart[1:4]) == "XML" {
					return true
				}
			}

			if h5.tokenLen > 5 {
				// IE <?import pseudo-tag
				if strings.ToUpper(strings.ReplaceAll(h5.tokenStart[:6], "\x00", "")) == "IMPORT" {
					return true
				}

				// XML Entity definition
				if strings.ToUpper(strings.ReplaceAll(h5.tokenStart[:6], "\x00", "")) == "ENTITY" {
					return true
				}
			}
		}
	}

	return false
}

func IsXSS(input string) bool {
	if isXss(input, html5FlagsDataState) ||
		isXss(input, html5FlagsValueNoQuote) ||
		isXss(input, html5FlagsValueSingleQuote) ||
		isXss(input, html5FlagsValueDoubleQuote) ||
		isXss(input, html5FlagsValueBackQuote) {
		return true
	}

	return false
}
