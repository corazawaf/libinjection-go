package libinjection

const (
	Html5TypeDataText = iota
	Html5TypeTagNameOpen
	Html5TypeTagNameClose
	Html5TypeTagNameSelfClose
	Html5TypeTagData
	Html5TypeTagClose
	Html5TypeAttrName
	Html5TypeAttrValue
	Html5TypeTagComment
	Html5TypeDocType
)

const (
	Html5FlagsDataState = iota
	Html5FlagsValueNoQuote
	Html5FlagsValueSingleQuote
	Html5FlagsValueDoubleQuote
	Html5FlagsValueBackQuote
)

type fnH5State func(*h5State)

type h5State struct {
	s          string
	len        int
	pos        int
	state      fnH5State
	tokenStart int
	tokenLen   int
	tokenType  int
}
