package libinjection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsXSS(t *testing.T) {
	examples := []string{
		"<script>alert(1);</script>",
		"><script>alert(1);</script>",
		"x ><script>alert(1);</script>",
		"' ><script>alert(1);</script>",
		"\"><script>alert(1);</script>",
		"red;</style><script>alert(1);</script>",
		"red;}</style><script>alert(1);</script>",
		"red;\"/><script>alert(1);</script>",
		"');}</style><script>alert(1);</script>",
		"onerror=alert(1)>",
		"x onerror=alert(1);>",
		"x' onerror=alert(1);>",
		"x\" onerror=alert(1);>",
		"<a href=\"javascript:alert(1)\">",
		"<a href='javascript:alert(1)'>",
		"<a href=javascript:alert(1)>",
		"<a href  =   javascript:alert(1); >",
		"<a href=\"  javascript:alert(1);\" >",
		"<a href=\"JAVASCRIPT:alert(1);\" >",
	}

	for _, example := range examples {
		if !IsXSS(example) {
			t.Errorf("[%s] is not XSS", example)
		}
	}
}

const (
	html5 = "html5"
	xss   = "xss"
)

var xssCount = 0

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
		xssCount++
		fmt.Println("FILE: (" + filename + ")")
		fmt.Println("INPUT: (" + data["--INPUT--"] + ")")
		fmt.Println("EXPECTED: (" + data["--EXPECTED--"] + ")")
		fmt.Println("GOT: (" + actual + ")")
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
			runXSSTest(t, data, p, html5)
		}
	}

	t.Log("False testing count: ", xssCount)
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
