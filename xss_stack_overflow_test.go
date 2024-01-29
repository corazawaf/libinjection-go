package libinjection

import (
	"testing"
)

func TestMemory(t *testing.T) {
	size := 10_000_000
	input := make([]byte, size)
	for i := range input {
		input[i] = '/'
	}

	// should not overflow the stack
	IsXSS(string(input))
}
