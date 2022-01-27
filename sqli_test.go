package libinjection

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestIsSQLi(t *testing.T) {
	result, fingerprint := IsSQLi("1�UNION�SELECT�2--")
	fmt.Println("=========result==========: ", result)
	fmt.Println("=======fingerprint=======: ", string(fingerprint[:]))
}

const (
	fingerprints = "fingerprints"
	folding      = "folding"
	tokens       = "tokens"
)

var sqliCount = 0

func printTokenString(t *sqliToken) string {
	out := ""
	if t.strOpen != 0 {
		out += string(t.strOpen)
	}
	out += string(t.val[:t.len])
	if t.strClose != 0 {
		out += string(t.strClose)
	}
	return out
}

func printToken(t *sqliToken) string {
	out := ""
	out += string(t.category)
	out += " "
	if t.category == 's' {
		out += printTokenString(t)
	} else if t.category == 'v' {
		vc := t.count
		if vc == 1 {
			out += "@"
		} else if vc == 2 {
			out += "@@"
		}
		out += printTokenString(t)
	} else {
		out += string(t.val[:t.len])
	}
	return strings.TrimSpace(out)
}

func getToken(state *sqliState, i int) *sqliToken {
	if i < 0 || i > maxTokens {
		panic("token got error!")
	}
	return &state.tokenVec[i]
}

func readTestData(filename string) map[string]string {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var (
		data  = make(map[string]string)
		state = ""
	)

	br := bufio.NewReaderSize(f, 8192)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		str := string(bytes.TrimSpace(line))
		if str == "--TEST--" || str == "--INPUT--" || str == "--EXPECTED--" {
			state = str
		} else {
			data[state] += str + "\n"
		}
	}
	data["--TEST--"] = strings.TrimSpace(data["--TEST--"])
	data["--INPUT--"] = strings.TrimSpace(data["--INPUT--"])
	data["--EXPECTED--"] = strings.TrimSpace(data["--EXPECTED--"])
	return data
}

func runSQLiTest(filename, flag string, sqliFlag int) {
	var (
		actual = ""
		data   = readTestData(filename)
		state  = new(sqliState)
	)

	sqliInit(state, data["--INPUT--"], sqliFlag)

	switch flag {
	case fingerprints:
		result, fingerprints := IsSQLi(data["--INPUT--"])
		if result {
			actual = string(fingerprints[:])
		}

	case folding:
		numTokens := state.fold()
		for i := 0; i < numTokens; i++ {
			actual += printToken(getToken(state, i)) + "\n"
		}

	case tokens:
		for state.tokenize() {
			actual += printToken(state.current) + "\n"
		}
	}

	actual = strings.TrimSpace(actual)
	if actual != data["--EXPECTED--"] {
		sqliCount += 1
		fmt.Println("FILE: (" + filename + ")")
		fmt.Println("INPUT: (" + data["--INPUT--"] + ")")
		fmt.Println("EXPECTED: (" + data["--EXPECTED--"] + ")")
		fmt.Println("GOT: (" + actual + ")")
	}
}

func TestSQLiDriver(t *testing.T) {
	baseDir := "./tests/"
	dir, err := ioutil.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, fi := range dir {
		if strings.Contains(fi.Name(), "-sqli-") {
			runSQLiTest(baseDir+fi.Name(), fingerprints, 0)
		} else if strings.Contains(fi.Name(), "-folding-") {
			runSQLiTest(baseDir+fi.Name(), folding, sqliFlagQuoteNone|sqliFlagSQLAnsi)
		} else if strings.Contains(fi.Name(), "-tokens_mysql-") {
			runSQLiTest(baseDir+fi.Name(), tokens, sqliFlagQuoteNone|sqliFlagSQLMysql)
		} else if strings.Contains(fi.Name(), "-tokens-") {
			runSQLiTest(baseDir+fi.Name(), tokens, sqliFlagQuoteNone|sqliFlagSQLAnsi)
		}
	}

	t.Log("False testing count: ", sqliCount)
}
