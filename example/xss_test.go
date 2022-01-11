package example

import (
	"strings"
	"testing"
)

func TestIsXSS(t *testing.T) {
	t.Log(strings.IndexByte("abc", 'b'))
}
