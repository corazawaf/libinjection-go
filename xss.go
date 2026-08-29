package libinjection

import (
	"strings"
	"sync"
)

var h5StatePool = sync.Pool{New: func() any { return new(h5State) }}

func isXSS(input string, flags int) bool {
	h5 := h5StatePool.Get().(*h5State)
	defer func() {
		*h5 = h5State{} // clear input/token references before returning to pool
		h5StatePool.Put(h5)
	}()
	h5.init(input, flags) // full reset then re-init
	return runXSS(h5)
}

// runXSS contains the detection loop; it is split out so isXSS can own the
// deferred pool Put while this function keeps the fast-return paths clean.
func runXSS(h5 *h5State) bool {
	attr := attributeTypeNone
	for h5.next() {
		if h5.tokenType != html5TypeAttrValue {
			attr = attributeTypeNone
		}

		detected, nextAttr := checkToken(h5, attr)
		if detected {
			return true
		}
		attr = nextAttr
	}

	return false
}

// checkToken reports whether the current token triggers XSS detection, and
// returns the attribute context to carry into the next token. An attribute
// name sets the context that its value is then judged against; every other
// token type leaves the incoming context alone.
func checkToken(h5 *h5State, attr int) (detected bool, nextAttr int) {
	switch h5.tokenType {
	case html5TypeDocType:
		return true, attr
	case html5TypeTagNameOpen:
		return isBlackTag(h5.tokenStart[:h5.tokenLen]), attr
	case html5TypeAttrName:
		return false, isBlackAttr(h5.tokenStart[:h5.tokenLen])
	case html5TypeAttrValue:
		return handleAttrValue(h5.tokenStart, h5.tokenLen, attr), attributeTypeNone
	case html5TypeTagComment:
		return handleTagComment(h5.tokenStart, h5.tokenLen), attr
	}
	return false, attr
}

// handleAttrValue reports whether an attribute value triggers XSS detection
// given the current attribute context (attr).
//
// IE6/7/8 may hide a full <script> or other black tag inside an attribute
// value under HTML5 parsing; see http://html5sec.org/#102.
func handleAttrValue(tokenStart string, tokenLen int, attr int) bool {
	switch attr {
	case attributeTypeNone:
		return false
	case attributeTypeBlack:
		return true
	case attributeTypeAttrURL:
		return isBlackURL(tokenStart[:tokenLen])
	case attributeTypeStyle:
		return true
	case attributeTypeAttrIndirect:
		// an attribute name is specified in a _value_
		return isBlackAttr(tokenStart[:tokenLen]) == attributeTypeBlack
	}
	return false
}

// hasCommentPrefix reports whether a comment token opens with an IE
// conditional comment ("[if") or an XML namespace declaration ("xml").
func hasCommentPrefix(token string) bool {
	if len(token) < 4 {
		return false
	}
	return strings.EqualFold(token[:3], "[if") || strings.EqualFold(token[:3], "xml")
}

// handleTagComment reports whether an HTML comment token triggers XSS
// detection (IE backtick terminator, IE conditional comments, XML namespace
// declarations, or IE import/entity pseudo-tags).
func handleTagComment(tokenStart string, tokenLen int) bool {
	// IE uses "`" as a tag ending byte.
	if strings.IndexByte(tokenStart[:tokenLen], '`') != -1 {
		return true
	}

	if hasCommentPrefix(tokenStart[:tokenLen]) {
		return true
	}

	// IE <?import pseudo-tag or XML Entity definition. The whole token is
	// normalized, not just its first 6 bytes: upperRemoveNulls drops NULs, so
	// an embedded NUL shifts the keyword past a fixed-width window.
	if tokenLen > 5 {
		var buf [6]byte
		n, _ := upperRemoveNulls(buf[:], tokenStart[:tokenLen])
		if n == 6 && (string(buf[:6]) == "IMPORT" || string(buf[:6]) == "ENTITY") {
			return true
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
