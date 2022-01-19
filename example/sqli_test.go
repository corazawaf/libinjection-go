package example

import (
	"fmt"
	"github.com/bxlxx/libinjection-go"
	"testing"
)

func TestIsSQLi(t *testing.T) {
	result, fingerprint := libinjection.IsSQLi("-1' and 1=1--")
	fmt.Println("=========result==========: ", result)
	fmt.Println("=======fingerprint=======: ", string(fingerprint[:]))
}
