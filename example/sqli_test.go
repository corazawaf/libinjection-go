package example

import (
	"fmt"
	"github.com/bxlxx/libinjection-go"
	"testing"
)

func TestIsSQLi(t *testing.T) {
	result, fingerprint := libinjection.IsSQLi("' OR '1'='1' --")
	fmt.Println("=========result==========: ", result)
	fmt.Println("=======fingerprint=======: ", fingerprint[:])
}
