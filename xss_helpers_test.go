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
		{
			name: "Test with onauxclick event handler",
			attr: "onauxclick",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onbeforematch event handler",
			attr: "onbeforematch",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onbegin event handler (SVG)",
			attr: "onbegin",
			want: attributeTypeBlack,
		},
		{
			name: "Test with oncommand event handler",
			attr: "oncommand",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onpagereveal event handler",
			attr: "onpagereveal",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onpageswap event handler",
			attr: "onpageswap",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onredraw event handler (SVG)",
			attr: "onredraw",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onrepeat event handler (SVG)",
			attr: "onrepeat",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onrepeatevent event handler (SVG)",
			attr: "onrepeatevent",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onscrollend event handler",
			attr: "onscrollend",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onscrollsnapchange event handler",
			attr: "onscrollsnapchange",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onscrollsnapchanging event handler",
			attr: "onscrollsnapchanging",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onwebkitassociateformcontrols event handler",
			attr: "onwebkitassociateformcontrols",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onwebkitautofillrequest event handler",
			attr: "onwebkitautofillrequest",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onwebkitmediasessionmetadatachanged event handler",
			attr: "onwebkitmediasessionmetadatachanged",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onwebkitshadowrootattached event handler",
			attr: "onwebkitshadowrootattached",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onwebkitwillrevealbottom event handler",
			attr: "onwebkitwillrevealbottom",
			want: attributeTypeBlack,
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
