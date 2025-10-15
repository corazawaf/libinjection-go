package libinjection

// No imports needed after optimization

type sqliToken struct {
	// position and length of token in original string
	pos int
	len int

	// count: in type 'v', used for number of opening '@', but maybe used in other contexts
	count int

	category byte
	strOpen  byte
	strClose byte
	val      string
}

const (
	maxTokens = 5
	tokenSize = 32
)

// Look forward for doubling of delimiter
//
// case 'foo' 'bar' -> foo' 'bar
//
// ending quote is not duplicated (i.e. escaped)
// since it's the wrong or EOL
func (t *sqliToken) parseStringCore(s string, length, pos, offset int, delimiter byte) int {
	// offset is to skip the perhaps first quote char
	searchStart := pos + offset

	if offset > 0 {
		// this is real quote
		t.strOpen = delimiter
	} else {
		// this was a simulated quote
		t.strOpen = byteNull
	}

	// Use direct byte search instead of string operations
	for {
		index := -1
		for i := searchStart; i < length; i++ {
			if s[i] == delimiter {
				index = i
				break
			}
		}

		if index == -1 {
			// string ended with no trailing quote
			// assign what we have
			t.assign(sqliTokenTypeString, pos+offset, length-pos-offset, s[pos+offset:])
			t.strClose = byteNull
			return length
		}

		// Check for backslash escaping
		if isBackslashEscaped(s[pos+offset : index]) {
			// keep going, move ahead one character
			searchStart = index + 1
			continue
		}

		// Check for double delimiter escaping
		if index+1 < length && s[index+1] == delimiter {
			// keep going, move ahead two characters
			searchStart = index + 2
			continue
		}

		// hey it's a normal string
		t.assign(sqliTokenTypeString, pos+offset, index-pos-offset, s[pos+offset:])
		t.strClose = delimiter
		return index + 1
	}
}

func (t *sqliToken) assign(tokenType byte, pos, length int, value string) {
	var last int
	if length < tokenSize {
		last = length
	} else {
		last = tokenSize - 1
	}

	t.category = tokenType
	t.pos = pos
	t.len = last
	// Avoid string slicing if possible
	if last == len(value) {
		t.val = value
	} else {
		t.val = value[:last]
	}
}

func (t *sqliToken) isUnaryOp() bool {
	if t.category != sqliTokenTypeOperator {
		return false
	}

	switch t.len {
	case 1:
		return t.val[0] == '+' || t.val[0] == '-' || t.val[0] == '!' || t.val[0] == '~'
	case 2:
		return t.val[0] == '!' && t.val[1] == '!'
	case 3:
		// Direct byte comparison instead of toUpperCmp to avoid allocation
		return (t.val[0] == 'N' || t.val[0] == 'n') &&
			(t.val[1] == 'O' || t.val[1] == 'o') &&
			(t.val[2] == 'T' || t.val[2] == 't')
	default:
		return false
	}
}

func (t *sqliToken) isArithmeticOp() bool {
	return t.category == sqliTokenTypeOperator && t.len == 1 &&
		(t.val[0] == '*' || t.val[0] == '/' || t.val[0] == '+' || t.val[0] == '-' || t.val[0] == '%')
}
