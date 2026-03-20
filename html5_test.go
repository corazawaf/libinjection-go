package libinjection

import "testing"

// h5TokenInfo captures the observable output of one tokenizer step.
type h5TokenInfo struct {
	typ int
	len int
}

// nextTokenInfos drives the h5State to collect all token type+length pairs from input.
func nextTokenInfos(input string, flags int) []h5TokenInfo {
	h := new(h5State)
	h.init(input, flags)
	var toks []h5TokenInfo
	for h.next() {
		toks = append(toks, h5TokenInfo{typ: h.tokenType, len: h.tokenLen})
	}
	return toks
}

// checkTokens fails the test if the actual token list doesn't match the expected list.
func checkTokens(t *testing.T, got []h5TokenInfo, want ...h5TokenInfo) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d tokens %v, want %d tokens %v", len(got), got, len(want), want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("token[%d]: got {typ=%d, len=%d}, want {typ=%d, len=%d}",
				i, got[i].typ, got[i].len, w.typ, w.len)
		}
	}
}

// TestSkipWhiteEOF exercises the return-byteEOF path in skipWhite,
// reached when the input ends with only whitespace after the tag name.
func TestSkipWhiteEOF(t *testing.T) {
	// "<div   " ends with spaces; stateBeforeAttributeName calls skipWhite
	// which runs out of input and returns byteEOF.
	// Expected: one TagNameOpen token for "div" (len=3).
	got := nextTokenInfos("<div   ", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagNameOpen, 3})
}

// TestStateBogusComment2Continue exercises the continue branch inside
// stateBogusComment2 that is taken when % is found but not followed by >.
func TestStateBogusComment2Continue(t *testing.T) {
	// "<%a%b>" – the first % is followed by 'a' (not >), so we continue;
	// no second % precedes >, so we reach EOF and emit the full remainder as a comment.
	// Expected: one TagComment token covering the 4 bytes "a%b>" (len=4).
	got := nextTokenInfos("<%a%b>", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 4})
}

// TestStateBogusComment2PctGT exercises the "ends in %>" happy path in
// stateBogusComment2.
func TestStateBogusComment2PctGT(t *testing.T) {
	// "<%foo%>" – % is followed by >, so the comment ends there.
	// Expected: one TagComment token for the 3-byte content "foo" (len=3).
	got := nextTokenInfos("<%foo%>", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 3})
}

// TestStateCommentNullBytes exercises the inner null-byte-skipping loop inside
// stateComment.
func TestStateCommentNullBytes(t *testing.T) {
	// "<!---\x00->" – after "<!--", the comment parser finds '-', skips the null byte,
	// finds '-' again, then '>'. The comment content between <!-- and --> is empty (len=0).
	got := nextTokenInfos("<!---\x00->", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 0})
}

// TestStateCommentNonDashBangContinue exercises the continue path in stateComment
// where a '-' is found but not followed by '-' or '!'.
func TestStateCommentNonDashBangContinue(t *testing.T) {
	// "<!--foo-bar-->" – the '-' before 'bar' is not followed by '-' or '!',
	// so the loop continues until '-->' is found.
	// Expected: one TagComment token for "foo-bar" (len=7).
	got := nextTokenInfos("<!--foo-bar-->", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 7})
}

// TestStateCommentNullBytesEOF exercises the EOF-after-null-bytes path in stateComment.
func TestStateCommentNullBytesEOF(t *testing.T) {
	// "<!---\x00\x00" – the null bytes after the leading '-' fill up to EOF,
	// triggering the EOF branch inside stateComment.
	// Expected: one TagComment token for the 3-byte remainder (len=3).
	got := nextTokenInfos("<!---\x00\x00", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 3})
}

// TestStateCommentDashBangNoGT exercises the continue path in stateComment when
// '-' or '!' is found but the following character is not '>'.
func TestStateCommentDashBangNoGT(t *testing.T) {
	// "<!--foo-!bar-->" – the '!' following '-' is not '>', so the inner loop
	// re-iterates and eventually finds '-->'.
	// Expected: one TagComment token for "foo-!bar" (len=8).
	got := nextTokenInfos("<!--foo-!bar-->", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 8})
}

// TestStateCommentEOFAfterDashNull exercises the EOF path in stateComment after
// a null byte is skipped and the offset reaches h.len.
func TestStateCommentEOFAfterDashNull(t *testing.T) {
	// "<!---\x00-" – after skipping the null byte, '-' is the last character,
	// so the offset reaches h.len and the EOF branch fires.
	// Expected: one TagComment token for the 3-byte remainder (len=3).
	got := nextTokenInfos("<!---\x00-", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 3})
}

// TestStateEndTagOpenEOF exercises the return-false EOF path in stateEndTagOpen.
func TestStateEndTagOpenEOF(t *testing.T) {
	// "</": stateTagOpen increments h.pos past '/', then stateEndTagOpen finds EOF.
	got := nextTokenInfos("</", html5FlagsDataState)
	if len(got) != 0 {
		t.Fatalf("expected no tokens for EOF after '</', got %v", got)
	}
}

// TestStateEndTagOpenGT exercises the stateData() call in stateEndTagOpen when
// the character after '</' is '>'.
func TestStateEndTagOpenGT(t *testing.T) {
	// "</>" – stateEndTagOpen sees '>' immediately → delegates to stateData,
	// which emits the remaining ">" as a DataText token (len=1).
	got := nextTokenInfos("</>", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeDataText, 1})
}

// TestStateEndTagOpenBogus exercises the stateBogusComment() call in stateEndTagOpen
// when the character after '</' is neither a letter nor '>'.
func TestStateEndTagOpenBogus(t *testing.T) {
	// "</0abc>" – '0' is not a letter, so stateEndTagOpen calls stateBogusComment,
	// which emits everything up to '>' as a TagComment token (len=4 for "0abc").
	got := nextTokenInfos("</0abc>", html5FlagsDataState)
	checkTokens(t, got, h5TokenInfo{html5TypeTagComment, 4})
}

// TestStateTagOpenNullChar exercises the ch==byteNull branch in stateTagOpen.
func TestStateTagOpenNullChar(t *testing.T) {
	// "<\x00div>" – null byte as first char after '<' is passed to stateTagName,
	// which emits a TagNameOpen token ("\x00div", len=4) followed by TagNameClose (len=1).
	got := nextTokenInfos("<\x00div>", html5FlagsDataState)
	checkTokens(t, got,
		h5TokenInfo{html5TypeTagNameOpen, 4},
		h5TokenInfo{html5TypeTagNameClose, 1},
	)
}

// TestStateTagOpenDefaultNonZeroPos exercises the default branch in stateTagOpen
// when h.pos > 0 (returns a DATA_TEXT token for the '<').
func TestStateTagOpenDefaultNonZeroPos(t *testing.T) {
	// "<1foo>" – '1' is not '!', '/', '?', '%', a letter, or null, and h.pos > 0,
	// so stateTagOpen emits a DataText token for the '<' (len=1), then stateData
	// emits the remaining "1foo>" as another DataText token (len=5).
	got := nextTokenInfos("<1foo>", html5FlagsDataState)
	checkTokens(t, got,
		h5TokenInfo{html5TypeDataText, 1},
		h5TokenInfo{html5TypeDataText, 5},
	)
}

// TestStateTagOpenDefaultZeroPos exercises the default branch in stateTagOpen when
// h.pos == 0, which calls stateData directly.
func TestStateTagOpenDefaultZeroPos(t *testing.T) {
	// Construct an h5State that starts directly in stateTagOpen with pos=0.
	// The first character '1' is not '!', '/', '?', '%', a letter, or null,
	// so the default branch fires. With h.pos == 0 it delegates to stateData().
	h := &h5State{}
	h.s = "1foo"
	h.len = 4
	h.pos = 0
	h.state = h.stateTagOpen

	result := h.next()
	if !result {
		t.Fatal("expected h.next() to return true")
	}
	if h.tokenType != html5TypeDataText {
		t.Fatalf("expected html5TypeDataText token, got %d", h.tokenType)
	}
	if h.tokenLen == 0 {
		t.Fatal("expected non-empty token")
	}
}

// TestStateAfterAttributeNameEOF exercises the byteEOF path in stateAfterAttributeName.
func TestStateAfterAttributeNameEOF(t *testing.T) {
	// "<div foo   " – attribute name followed by spaces then EOF.
	// stateAttributeName sets state to stateAfterAttributeName; skipWhite runs out.
	// Expected: TagNameOpen "div" (len=3) + AttrName "foo" (len=3).
	got := nextTokenInfos("<div foo   ", html5FlagsDataState)
	checkTokens(t, got,
		h5TokenInfo{html5TypeTagNameOpen, 3},
		h5TokenInfo{html5TypeAttrName, 3},
	)
}

// TestStateAfterAttributeNameSlash exercises the byteSlash path in stateAfterAttributeName.
func TestStateAfterAttributeNameSlash(t *testing.T) {
	// "<div foo />" – '/' after whitespace following attribute name triggers
	// stateSelfClosingStartTag, which emits a TagNameSelfClose token (len=2 for "/>").
	// Expected: TagNameOpen (len=3), AttrName (len=3), TagNameSelfClose (len=2).
	got := nextTokenInfos("<div foo />", html5FlagsDataState)
	checkTokens(t, got,
		h5TokenInfo{html5TypeTagNameOpen, 3},
		h5TokenInfo{html5TypeAttrName, 3},
		h5TokenInfo{html5TypeTagNameSelfClose, 2},
	)
}

// TestStateAfterAttributeNameGT exercises the byteGT path in stateAfterAttributeName.
func TestStateAfterAttributeNameGT(t *testing.T) {
	// "<div foo >" – '>' after whitespace following attribute name calls stateTagNameClose.
	// Expected: TagNameOpen (len=3), AttrName (len=3), TagNameClose (len=1).
	got := nextTokenInfos("<div foo >", html5FlagsDataState)
	checkTokens(t, got,
		h5TokenInfo{html5TypeTagNameOpen, 3},
		h5TokenInfo{html5TypeAttrName, 3},
		h5TokenInfo{html5TypeTagNameClose, 1},
	)
}

// TestStateBeforeAttributeValueEOF exercises the byteEOF path in stateBeforeAttributeValue.
func TestStateBeforeAttributeValueEOF(t *testing.T) {
	// "<div href=   " – attribute '=' then spaces then EOF; skipWhite returns EOF,
	// and stateBeforeAttributeValue returns false.
	// Expected: TagNameOpen "div" (len=3) + AttrName "href" (len=4).
	got := nextTokenInfos("<div href=   ", html5FlagsDataState)
	checkTokens(t, got,
		h5TokenInfo{html5TypeTagNameOpen, 3},
		h5TokenInfo{html5TypeAttrName, 4},
	)
}

// TestStateBeforeAttributeNameSlashContinue exercises the continue path in
// stateBeforeAttributeName when '/' is found but not followed by '>'.
func TestStateBeforeAttributeNameSlashContinue(t *testing.T) {
	// "<div / foo>" – '/' is not followed by '>', so the loop continues and
	// "foo" is parsed as an attribute name followed by '>'.
	// Expected: TagNameOpen (len=3), AttrName (len=3), TagNameClose (len=1).
	got := nextTokenInfos("<div / foo>", html5FlagsDataState)
	checkTokens(t, got,
		h5TokenInfo{html5TypeTagNameOpen, 3},
		h5TokenInfo{html5TypeAttrName, 3},
		h5TokenInfo{html5TypeTagNameClose, 1},
	)
}
