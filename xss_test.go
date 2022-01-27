package libinjection

import (
	"fmt"
	"testing"
)

func TestIsXSS(t *testing.T) {
	fmt.Println("result: ", IsXSS("<script>alert(1)</script>"))
}
