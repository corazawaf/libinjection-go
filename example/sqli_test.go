package example

import (
	"fmt"
	"libinjection"
	"testing"
)

func TestIsSQLi(t *testing.T) {
	result, fingerprint := libinjection.IsSQLi(" '1'='1' --")
	fmt.Println("=========result==========: ", result)
	fmt.Println("=======fingerprint=======: ", fingerprint[:])
}
