package libinjection

import "testing"

// Browsers strip TAB, LF and CR anywhere in a URL (WHATWG URL parsing), so
// they must not hide a dangerous scheme from isBlackURL.
func TestIsXSSURLWhitespace(t *testing.T) {
	// The cases mirror test_xss_url_whitespace() in the C library's
	// src/test_error_handling.c, which is where the fix originates; keep them
	// in step so the two ports do not drift.
	attacks := map[string]string{
		"href with raw TAB":                               "<a href=\"ja\tvascript:alert(1)\">x</a>",
		"href with raw CR":                                "<a href=\"ja\rvascript:alert(1)\">x</a>",
		"href with &#9;":                                  "<a href=\"ja&#9;vascript:alert(1)\">x</a>",
		"href with &#x09;":                                "<a href=\"ja&#x09;vascript:alert(1)\">x</a>",
		"href with &#13;":                                 "<a href=\"ja&#13;vascript:alert(1)\">x</a>",
		"href with &#10;":                                 "<a href=\"ja&#10;vascript:alert(1)\">x</a>",
		"href with &Tab;":                                 "<a href=\"ja&Tab;vascript:alert(1)\">x</a>",
		"href with &NewLine;":                             "<a href=\"ja&NewLine;vascript:alert(1)\">x</a>",
		"href with leading &Tab;":                         "<a href=\"&Tab;javascript:alert(1)\">x</a>",
		"href with leading &NewLine;":                     "<a href='&NewLine;javascript:alert(1)'>Click Me</a>",
		"form action with raw TAB":                        "<form action=\"ja\tvascript:alert(1)\">",
		"button formaction with &#9;":                     "<button formaction=\"ja&#9;vascript:alert(1)\">x</button>",
		"vbscript with &#9;":                              "<a href=\"vb&#9;script:x\">x</a>",
		"data with &#9;":                                  "<a href=\"da&#9;ta:text/html,x\">x</a>",
		"vbscript with &NewLine;":                         "<a href=\"vb&NewLine;script:x\">x</a>",
		"view-source with &Tab;":                          "<a href=\"view-so&Tab;urce:http://x\">x</a>",
		"consecutive &Tab;&Tab;":                          "<a href=\"ja&Tab;&Tab;vascript:alert(1)\">x</a>",
		"&#9; mixed with &#0;":                            "<a href=\"ja&#9;&#0;vascript:alert(1)\">x</a>",
		"single-quoted href with &Tab;":                   "<a href='ja&Tab;vascript:alert(1)'>x</a>",
		"unquoted href with &Tab;":                        "<a href=ja&Tab;vascript:alert(1)>x</a>",
		"complete prefix then a reference above 0xFF":     "<a href=\"java&#330;\">x</a>",
		"complete prefix then a large hex reference":      "<a href=\"data&#x1004A;:\">x</a>",
		"complete prefix then a hex reference above 0xFF": "<a href=\"vbscript&#x100;\">x</a>",
		"complete prefix then TAB":                        "<a href=\"java\t\">x</a>",
		"raw TAB mixed with &#9; inside view-source":      "<a href=\"view-so\t&#9;urce:http://x\">x</a>",
		"backtick href with &Tab;":                        "<a href=`ja&Tab;vascript:alert(1)`>x</a>",
	}
	for name, in := range attacks {
		t.Run(name, func(t *testing.T) {
			if !IsXSS(in) {
				t.Errorf("IsXSS(%q) = false, want true", in)
			}
		})
	}

	benign := map[string]string{
		"http URL with &#9; and &Tab;":            "<a href=\"http://ex&#9;ample.com/&Tab;\">x</a>",
		"lowercase &tab; is not decoded":          "<a href=\"ja&tab;vascript:alert(1)\">x</a>",
		"&Tab without semicolon is not decoded":   "<a href=\"ja&Tabvascript:alert(1)\">x</a>",
		"raw VT is not stripped":                  "<a href=\"ja\vvascript:alert(1)\">x</a>",
		"&#330; is not 'J' (no 8-bit truncation)": "<a href=\"&#330;avascript:alert(1)\">x</a>",
		"&Tab cut off at end of input":            "<a href=\"ja&Tab",
		"&NewLin cut off at end of input":         "<a href=\"ja&NewLin",
		"truncated &NewLin before the scheme":     "<a href=\"ja&NewLinvascript:alert(1)\">x</a>",
		"&Tab; inside http is no scheme":          "<a href=\"ht&Tab;tp://x\">x</a>",
	}
	for name, in := range benign {
		t.Run(name, func(t *testing.T) {
			if IsXSS(in) {
				t.Errorf("IsXSS(%q) = true, want false", in)
			}
		})
	}
}
