package libinjection

const (
	sqliFlagNone        = 0
	sqliFlagQuoteNone   = 1
	sqliFlagQuoteSingle = 2
	sqliFlagQuoteDouble = 4
	sqliFlagSQLAnsi     = 8
	sqliFlagSQLMysql    = 16
)

const (
	lookupWord        = 1
	lookupType        = 2
	lookupOperator    = 3
	lookupFingerprint = 4
)

const (
	byteNull   uint8 = '\000'
	byteSingle       = '\''
	byteDouble       = '"'
	byteTick         = '`'
)

const (
	sqliTokenTypeNone             byte = 0
	sqliTokenTypeKeyword          byte = 'k'
	sqliTokenTypeUnion            byte = 'U'
	sqliTokenTypeGroup            byte = 'B'
	sqliTokenTypeExpression       byte = 'E'
	sqliTokenTypeSQLType          byte = 't'
	sqliTokenTypeFunction         byte = 'f'
	sqliTokenTypeBareWord         byte = 'n'
	sqliTokenTypeNumber           byte = '1'
	sqliTokenTypeVariable         byte = 'v'
	sqliTokenTypeString           byte = 's'
	sqliTokenTypeOperator         byte = 'o'
	sqliTokenTypeLogicOperator    byte = '&'
	sqliTokenTypeComment          byte = 'c'
	sqliTokenTypeCollate          byte = 'A'
	sqliTokenTypeLeftParenthesis  byte = '('
	sqliTokenTypeRightParenthesis byte = ')'
	sqliTokenTypeLeftBrace        byte = '{'
	sqliTokenTypeRightBrace       byte = '}'
	sqliTokenTypeDot              byte = '.'
	sqliTokenTypeComma            byte = ','
	sqliTokenTypeColon            byte = ':'
	sqliTokenTypeSemiColon        byte = ';'
	sqliTokenTypeTSQL             byte = 'T'
	sqliTokenTypeUnknown          byte = '?'
	sqliTokenTypeEvil             byte = 'X'
	sqliTokenTypeFingerprint      byte = 'F'
	sqliTokenTypeBackslash        byte = '\\'
)
