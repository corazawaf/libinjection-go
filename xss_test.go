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
		// HTML5 event handler attributes (standard and experimental)
		{input: "<tag onauxclick=alert(1)>", isXSS: true},
		{input: "<tag onbeforematch=alert(1)>", isXSS: true},
		{input: "<tag onbegin=alert(1)>", isXSS: true},  // SVG animation
		{input: "<tag oncommand=alert(1)>", isXSS: true},
		{input: "<tag onpagereveal=alert(1)>", isXSS: true},        // View Transitions API
		{input: "<tag onpageswap=alert(1)>", isXSS: true},          // View Transitions API
		{input: "<tag onredraw=alert(1)>", isXSS: true},            // SVG
		{input: "<tag onrepeat=alert(1)>", isXSS: true},            // SVG animation
		{input: "<tag onrepeatevent=alert(1)>", isXSS: true},       // SVG animation
		{input: "<tag onscrollend=alert(1)>", isXSS: true},
		{input: "<tag onscrollsnapchange=alert(1)>", isXSS: true},  // CSS Scroll Snap
		{input: "<tag onscrollsnapchanging=alert(1)>", isXSS: true}, // CSS Scroll Snap
		{input: "<tag onwebkitassociateformcontrols=alert(1)>", isXSS: true}, // WebKit-specific
		{input: "<tag onwebkitautofillrequest=alert(1)>", isXSS: true},       // WebKit-specific
		{input: "<tag onwebkitmediasessionmetadatachanged=alert(1)>", isXSS: true}, // WebKit-specific
		{input: "<tag onwebkitshadowrootattached=alert(1)>", isXSS: true},          // WebKit-specific
		{input: "<tag onwebkitwillrevealbottom=alert(1)>", isXSS: true}, // WebKit-specific
		// Browser-specific event handlers (IE, Firefox, Edge)
		{input: "<tag onafterscriptexecute=alert(1)>", isXSS: true},  // Firefox
		{input: "<tag onafterupdate=alert(1)>", isXSS: true},         // IE
		{input: "<tag onbeforeactivate=alert(1)>", isXSS: true},      // IE
		{input: "<tag onbeforedeactivate=alert(1)>", isXSS: true},    // IE
		{input: "<tag onbeforeeditfocus=alert(1)>", isXSS: true},     // IE
		{input: "<tag onbeforescriptexecute=alert(1)>", isXSS: true}, // Firefox
		{input: "<tag onbeforeupdate=alert(1)>", isXSS: true},        // IE
		{input: "<tag oncellchange=alert(1)>", isXSS: true},          // IE
		{input: "<tag ondatasetchanged=alert(1)>", isXSS: true},      // IE
		{input: "<tag ondatasetcomplete=alert(1)>", isXSS: true},     // IE
		{input: "<tag ondeactivate=alert(1)>", isXSS: true},          // IE
		{input: "<tag onerrorupdate=alert(1)>", isXSS: true},         // IE
		{input: "<tag onfilterchange=alert(1)>", isXSS: true},        // IE
		{input: "<tag onlayoutcomplete=alert(1)>", isXSS: true},      // IE
		{input: "<tag onlosecapture=alert(1)>", isXSS: true},         // IE
		{input: "<tag onmozfullscreenchange=alert(1)>", isXSS: true}, // Firefox
		{input: "<tag onmozfullscreenerror=alert(1)>", isXSS: true},  // Firefox
		{input: "<tag onmozpointerlockchange=alert(1)>", isXSS: true}, // Firefox
		{input: "<tag onmozpointerlockerror=alert(1)>", isXSS: true}, // Firefox
		{input: "<tag onmsfullscreenchange=alert(1)>", isXSS: true},  // IE/Edge
		{input: "<tag onmsfullscreenerror=alert(1)>", isXSS: true},   // IE/Edge
		{input: "<tag onpropertychange=alert(1)>", isXSS: true},      // IE
		{input: "<tag onresizeend=alert(1)>", isXSS: true},           // IE
		{input: "<tag onresizestart=alert(1)>", isXSS: true},         // IE
		{input: "<tag onrowenter=alert(1)>", isXSS: true},            // IE
		{input: "<tag onrowexit=alert(1)>", isXSS: true},             // IE
		{input: "<tag onrowsdelete=alert(1)>", isXSS: true},          // IE
		{input: "<tag onrowsinserted=alert(1)>", isXSS: true},        // IE
		// Payload sample from https://github.com/payloadbox/xss-payload-list
		{input: "<HTML xmlns:xss><?import namespace=\"xss\" implementation=\"%(htc)s\"><xss:xss>XSS</xss:xss></HTML>\"\"\",\"XML namespace.\"),(\"\"\"<XML ID=\"xss\"><I><B>&lt;IMG SRC=\"javas<!-- -->cript:javascript:alert(1)\"&gt;</B></I></XML><SPAN DATASRC=\"#xss\" DATAFLD=\"B\" DATAFORMATAS=\"HTML\"></SPAN>", isXSS: true},
		// True negatives
		{input: "myvar=onfoobar==", isXSS: false},
		{input: "onY29va2llcw==", isXSS: false}, // base64 encoded "thisisacookie", prefixed by "on"
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
