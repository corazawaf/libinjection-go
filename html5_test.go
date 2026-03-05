package libinjection

import "testing"

// nextTokens drives the h5State to collect all tokens from the given input
// with the given initial flag.
func nextTokens(input string, flags int) []int {
	h := new(h5State)
	h.init(input, flags)
	var types []int
	for h.next() {
		types = append(types, h.tokenType)
	}
	return types
}

// TestSkipWhiteEOF exercises the return-byteEOF path in skipWhite,
// reached when the input ends with only whitespace after the tag name.
func TestSkipWhiteEOF(t *testing.T) {
	// "<div   " ends with spaces; stateBeforeAttributeName calls skipWhite
	// which runs out of input and returns byteEOF.
	tokens := nextTokens("<div   ", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateBogusComment2Continue exercises the continue branch inside
// stateBogusComment2 that is taken when % is found but not followed by >.
func TestStateBogusComment2Continue(t *testing.T) {
	// "<%a%b>" – the first % is followed by 'a' (not >), so we continue;
	// the second % is followed by... well, due to the search using h.pos
	// we eventually hit EOF.
	tokens := nextTokens("<%a%b>", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateBogusComment2PctGT exercises the "ends in %>" happy path in
// stateBogusComment2.
func TestStateBogusComment2PctGT(t *testing.T) {
	tokens := nextTokens("<%foo%>", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateCommentNullBytes exercises the inner null-byte-skipping loop inside
// stateComment.
func TestStateCommentNullBytes(t *testing.T) {
	// "<!---\x00->" – after "<!--", the comment parser finds '-', then skips
	// the null byte (\x00), then finds '-' again, then '>'.
	tokens := nextTokens("<!---\x00->", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateCommentNonDashBangContinue exercises the continue path in stateComment
// where a '-' is found but not followed by '-' or '!'.
func TestStateCommentNonDashBangContinue(t *testing.T) {
	// "<!--foo-bar-->" – the '-' before 'bar' is not followed by '-' or '!',
	// so the loop continues.
	tokens := nextTokens("<!--foo-bar-->", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateCommentNullBytesEOF exercises the EOF-after-null-bytes path in stateComment.
func TestStateCommentNullBytesEOF(t *testing.T) {
	// "<!---\x00\x00" – the null bytes after the leading '-' fill up to EOF,
	// triggering the second EOF branch inside stateComment.
	tokens := nextTokens("<!---\x00\x00", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateCommentDashBangNoGT exercises the continue path in stateComment when
// '-' or '!' is found but the following character is not '>'.
func TestStateCommentDashBangNoGT(t *testing.T) {
	// "<!--foo-!bar-->" – '-' at position 7 is followed by '!', which is not '>';
	// the inner loop re-iterates and eventually finds '-->'.
	tokens := nextTokens("<!--foo-!bar-->", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateCommentEOFAfterDashNull exercises the EOF path in stateComment after
// a null byte is skipped and the offset equals h.len.
func TestStateCommentEOFAfterDashNull(t *testing.T) {
	// "<!---\x00-" – after "<!--" (4 bytes) the sequence is '-' at index 4,
	// null at index 5, then '-' at index 6 (the last byte).
	// After skipping the null (offset=2 within stateComment) and confirming ch=='-',
	// offset becomes 3, which equals h.len; the EOF branch fires.
	tokens := nextTokens("<!---\x00-", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateEndTagOpenEOF exercises the return-false EOF path in stateEndTagOpen.
func TestStateEndTagOpenEOF(t *testing.T) {
	// "</": stateTagOpen increments h.pos past '/', then stateEndTagOpen finds EOF.
	tokens := nextTokens("</", html5FlagsDataState)
	// The only result is false (no token), so len(tokens)==0 is expected here.
	_ = tokens
}

// TestStateEndTagOpenGT exercises the stateData() call in stateEndTagOpen when
// the character after '</' is '>'.
func TestStateEndTagOpenGT(t *testing.T) {
	// "</>" – stateEndTagOpen sees '>' immediately → calls stateData.
	tokens := nextTokens("</>", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateEndTagOpenBogus exercises the stateBogusComment() call in stateEndTagOpen
// when the character after '</' is neither a letter nor '>'.
func TestStateEndTagOpenBogus(t *testing.T) {
	// "</0abc>" – '0' is not a letter, so stateEndTagOpen calls stateBogusComment.
	tokens := nextTokens("</0abc>", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateTagOpenNullChar exercises the ch==byteNull branch in stateTagOpen.
func TestStateTagOpenNullChar(t *testing.T) {
	// "<\x00div>" – null byte as first char after '<' goes through stateTagName.
	tokens := nextTokens("<\x00div>", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateTagOpenDefaultNonZeroPos exercises the default branch in stateTagOpen
// when h.pos > 0 (returns a DATA_TEXT token for the '<').
func TestStateTagOpenDefaultNonZeroPos(t *testing.T) {
	// "<1foo>" – '1' is not '!', '/', '?', '%', a letter, or null, and h.pos > 0.
	tokens := nextTokens("<1foo>", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateTagOpenDefaultZeroPos exercises the default branch in stateTagOpen when
// h.pos == 0, which calls stateData directly.
func TestStateTagOpenDefaultZeroPos(t *testing.T) {
	// Construct an h5State that starts directly in stateTagOpen with pos=0.
	h := &h5State{}
	h.s = "1foo"
	h.len = 4
	h.pos = 0
	h.state = h.stateTagOpen
	h.next()
	// No panic means the branch was executed.
}

// TestStateAfterAttributeNameEOF exercises the byteEOF path in stateAfterAttributeName.
func TestStateAfterAttributeNameEOF(t *testing.T) {
	// "<div foo   " – attribute name followed by spaces then EOF.
	// stateAttributeName sets state to stateAfterAttributeName; skipWhite runs out.
	tokens := nextTokens("<div foo   ", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateAfterAttributeNameSlash exercises the byteSlash path in stateAfterAttributeName.
func TestStateAfterAttributeNameSlash(t *testing.T) {
	// "<div foo />" – '/' after whitespace following attribute name.
	tokens := nextTokens("<div foo />", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateAfterAttributeNameGT exercises the byteGT path in stateAfterAttributeName.
func TestStateAfterAttributeNameGT(t *testing.T) {
	// "<div foo >" – '>' after whitespace following attribute name.
	tokens := nextTokens("<div foo >", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateBeforeAttributeValueEOF exercises the byteEOF path in stateBeforeAttributeValue.
func TestStateBeforeAttributeValueEOF(t *testing.T) {
	// "<div href=   " – attribute '=' then spaces then EOF; skipWhite returns EOF.
	tokens := nextTokens("<div href=   ", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}

// TestStateBeforeAttributeNameSlashContinue exercises the continue path in
// stateBeforeAttributeName when '/' is found but not followed by '>'.
func TestStateBeforeAttributeNameSlashContinue(t *testing.T) {
	// "<div / foo>" – '/' is not followed by '>', so the loop continues.
	tokens := nextTokens("<div / foo>", html5FlagsDataState)
	if len(tokens) == 0 {
		t.Fatal("expected at least one token")
	}
}
