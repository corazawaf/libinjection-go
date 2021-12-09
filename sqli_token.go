package libinjection

type sqliToken struct {
	// position and length of token in original string
	pos int
	len int

	// count: in type 'v', used for number of opening '@', but maybe used in other contexts
	count int

	category byte
	strOpen  byte
	strClose byte
	val      [32]byte
}

const (
	maxTokens = 5
	tokenSize = 32
)

// Look forward for doubling of delimiter
//
// case 'foo''bar' -> foo''bar
//
// ending quote is not duplicated (i.e. escaped)
// since it's the wrong or EOL
func (t *sqliToken) parseStringCore(s string, length, pos, offset int, delimiter byte) int {
	// offset is to skip the perhaps first quote char
	str := ""
	flag := 0
	for i := 0; i < length-pos-offset; i++ {
		if s[i+pos+offset] == delimiter {
			str = s[i+pos+offset:]
			flag = i + pos + offset
			break
		}
	}

	if offset > 0 {
		// this is real quote
		t.strOpen = delimiter
	} else {
		// this was a simulated quote
		t.strOpen = byteNull
	}

	for {
		if len(str) == 0 {
			// string ended with no trailing quote
			// assign what we have
			t.assign(sqliTokenTypeString, pos+offset, length-pos-offset, s[pos+offset:])
			t.strClose = byteNull
			return length
		} else if isBackslashEscaped(s[pos+offset : flag]) {
			// keep going, move ahead one character
			str = str[1:]
			continue
		} else if isDoubleDelimiterEscaped(str) {
			// keep going, move ahead two characters
			str = str[2:]
			continue
		} else {
			// hey it's a normal string
			t.assign(sqliTokenTypeString, pos+offset, len(s[pos+offset:])-len(str), s[pos+offset:])
			t.strClose = delimiter
			return len(s) - len(str) + 1
		}
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
	for i := 0; i < last; i++ {
		t.val[i] = value[i]
	}
	t.val[last] = byteNull
}

func (t *sqliToken) assignByte(tokenType byte, pos, len int, value byte) {
	t.category = tokenType
	t.pos = pos
	t.len = 1
	t.val[0] = value
	t.val[1] = byteNull
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
		return toUpperCmp("NOT", string(t.val[:3])) == 0
	default:
		return false
	}
}

func (t *sqliToken) isArithmeticOp() bool {
	return t.category == sqliTokenTypeOperator && t.len == 1 &&
		(t.val[0] == '*' || t.val[0] == '/' || t.val[0] == '+' || t.val[0] == '-')
}
