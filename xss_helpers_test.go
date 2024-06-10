package libinjection

import (
	"testing"
)

func TestIsBlackAttr(t *testing.T) {
	tests := []struct {
		name string
		attr string
		want int
	}{
		{
			name: "Test with black attribute",
			attr: "xmlns",
			want: attributeTypeBlack,
		},
		{
			name: "Test with non-black attribute",
			attr: "class",
			want: attributeTypeNone,
		},
		{
			name: "Test with JavaScript event handler",
			attr: "onclick",
			want: attributeTypeBlack,
		},
		{
			name: "Test with short attribute",
			attr: "a",
			want: attributeTypeNone,
		},
		{
			name: "Test with long null attribute that will be stripped",
			attr: "a\x00\x00\x00\x00\x00",
			want: attributeTypeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBlackAttr(tt.attr); got != tt.want {
				t.Errorf("isBlackAttr() = %v, want %v", got, tt.want)
			}
		})
	}
}
