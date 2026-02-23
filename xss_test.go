package libinjection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Examples can be read at https://portswigger.net/web-security/cross-site-scripting/cheat-sheet
func TestIsXSS(t *testing.T) {
	examples := []struct {
		input string
		isXSS bool
	}{
		// True positives
		{input: "<script>alert(1);</script>", isXSS: true},
		{input: "><script>alert(1);</script>", isXSS: true},
		{input: "x ><script>alert(1);</script>", isXSS: true},
		{input: "' ><script>alert(1);</script>", isXSS: true},
		{input: "\"><script>alert(1);</script>", isXSS: true},
		{input: "red;</style><script>alert(1);</script>", isXSS: true},
		{input: "red;}</style><script>alert(1);</script>", isXSS: true},
		{input: "red;\"/><script>alert(1);</script>", isXSS: true},
		{input: "');}</style><script>alert(1);</script>", isXSS: true},
		{input: "onerror=alert(1)>", isXSS: true},
		{input: "x onerror=alert(1);>", isXSS: true},
		{input: "x' onerror=alert(1);>", isXSS: true},
		{input: "x\" onerror=alert(1);>", isXSS: true},
		{input: "<a href=\"javascript:alert(1)\">", isXSS: true},
		{input: "<a href='javascript:alert(1)'>", isXSS: true},
		{input: "<a href=javascript:alert(1)>", isXSS: true},
		{input: "<a href  =   javascript:alert(1); >", isXSS: true},
		{input: "<a href=\"  javascript:alert(1);\" >", isXSS: true},
		{input: "<a href=\"JAVASCRIPT:alert(1);\" >", isXSS: true},
		{input: "<style>@keyframes x{}</style><xss style=\"animation-name:x\" onanimationstart=\"alert(1)\"></xss>", isXSS: true},
		{input: "<noembed><img title=\"</noembed><img src onerror=alert(1)>\"></noembed>", isXSS: true},
		{input: "javascript:/*--></title></style></textarea></script></xmp><svg/onload='+/\"/+/onmouseover=1/+/[*/[]/+alert(1)//'>", isXSS: true}, // polyglot payload
		{input: "<xss class=progress-bar-animated onanimationstart=alert(1)>", isXSS: true},
		{input: "<button popovertarget=x>Click me</button><xss ontoggle=alert(1) popover id=x>XSS</xss>", isXSS: true},
		// Payload sample from https://github.com/payloadbox/xss-payload-list
		{input: "<HTML xmlns:xss><?import namespace=\"xss\" implementation=\"%(htc)s\"><xss:xss>XSS</xss:xss></HTML>\"\"\",\"XML namespace.\"),(\"\"\"<XML ID=\"xss\"><I><B>&lt;IMG SRC=\"javas<!-- -->cript:javascript:alert(1)\"&gt;</B></I></XML><SPAN DATASRC=\"#xss\" DATAFLD=\"B\" DATAFORMATAS=\"HTML\"></SPAN>", isXSS: true},
		// Events from issue #70 - previously undetected event handlers
		{input: "<img onauxclick=alert(1)>", isXSS: true},
		{input: "<img onpagereveal=alert(1)>", isXSS: true},
		{input: "<img onpageswap=alert(1)>", isXSS: true},
		{input: "<img onscrollsnapchange=alert(1)>", isXSS: true},
		{input: "<img onscrollsnapchanging=alert(1)>", isXSS: true},
		// XML comment detection (tokenLen must be > 3 to reach this check)
		{input: "<!--xml -->", isXSS: true},
		{input: "<!--xmlfoo-->", isXSS: true},
		{input: "<!--xml:namespace-->", isXSS: true},
		{input: "<!--XML -->", isXSS: true},
		// True negatives
		{input: "<!--xml-->", isXSS: false},  // tokenLen=3, doesn't reach XML check
		{input: "<!--?xml -->", isXSS: false}, // "xml" not at start of token
		{input: "<!--axml -->", isXSS: false}, // "xml" not at start of token
		{input: "myvar=onfoobar==", isXSS: false},
		{input: "onY29va2llcw==", isXSS: false}, // base64 encoded "thisisacookie", prefixed by "on"
		// False positives from issue #46 - URLs containing black scheme names in path
		{input: `=<a href="https://data">`, isXSS: false},
		{input: `<a href="https://github.com/Simbiat/database">`, isXSS: false},
	}

	for _, example := range examples {
		if res := IsXSS(example.input); res != example.isXSS {
			t.Errorf("[%s] wanted: %t, got %t", example.input, example.isXSS, res)
		}
	}
}

const (
	html5 = "html5"
	xss   = "xss"
)

func h5TypeToString(h5Type int) string {
	switch h5Type {
	case html5TypeDataText:
		return "DATA_TEXT"
	case html5TypeTagNameOpen:
		return "TAG_NAME_OPEN"
	case html5TypeTagNameClose:
		return "TAG_NAME_CLOSE"
	case html5TypeTagNameSelfClose:
		return "TAG_NAME_SELFCLOSE"
	case html5TypeTagData:
		return "TAG_DATA"
	case html5TypeTagClose:
		return "TAG_CLOSE"
	case html5TypeAttrName:
		return "ATTR_NAME"
	case html5TypeAttrValue:
		return "ATTR_VALUE"
	case html5TypeTagComment:
		return "TAG_COMMENT"
	case html5TypeDocType:
		return "DOCTYPE"
	default:
		return ""
	}
}

func printHTML5Token(h *h5State) string {
	return fmt.Sprintf("%s,%d,%s",
		h5TypeToString(h.tokenType),
		h.tokenLen,
		h.tokenStart[:h.tokenLen])
}

func runXSSTest(t testing.TB, data map[string]string, filename, flag string) {
	t.Helper()
	var (
		actual = ""
	)

	switch flag {
	case xss:

	case html5:
		h5 := new(h5State)
		h5.init(data["--INPUT--"], html5FlagsDataState)

		for h5.next() {
			actual += printHTML5Token(h5) + "\n"
		}
	}

	actual = strings.TrimSpace(actual)
	if actual != data["--EXPECTED--"] {
		t.Errorf("FILE: (%s)\nINPUT: (%s)\nEXPECTED: (%s)\nGOT: (%s)\n",
			filename, data["--INPUT--"], data["--EXPECTED--"], actual)
	}
}

func TestXSSDriver(t *testing.T) {
	baseDir := "./tests/"
	dir, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range dir {
		p := filepath.Join(baseDir, fi.Name())
		data := readTestData(p)
		if strings.Contains(fi.Name(), "-html5-") {
			t.Run(fi.Name(), func(t *testing.T) {
				runXSSTest(t, data, p, html5)
			})
		}
	}
}

type testCaseXSS struct {
	name string
	data map[string]string
}

func BenchmarkXSSDriver(b *testing.B) {
	baseDir := "./tests/"
	dir, err := os.ReadDir(baseDir)
	if err != nil {
		b.Fatal(err)
	}

	cases := struct {
		html5 []testCaseXSS
	}{}

	for _, fi := range dir {
		p := filepath.Join(baseDir, fi.Name())
		data := readTestData(p)
		tc := testCaseXSS{
			name: fi.Name(),
			data: data,
		}
		switch {
		case strings.Contains(fi.Name(), "-html5-"):
			cases.html5 = append(cases.html5, tc)
		default:
		}
	}

	b.Run("html5", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for _, tc := range cases.html5 {
				tt := tc
				runXSSTest(b, tt.data, tt.name, html5)
			}
		}
	})
}

func TestXSS(t *testing.T) {
	tests := []struct {
		input string
		isXSS bool
	}{
		{
			input: "href=&#",
			isXSS: false,
		},
		{
			input: "href=&#X",
			isXSS: false,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.input, func(t *testing.T) {
			if want, have := tt.isXSS, IsXSS(tt.input); want != have {
				t.Errorf("want %v, have %v", want, have)
			}
		})
	}
}
