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
		// Browser-specific event handlers
		{
			name: "Test with onafterscriptexecute event handler (Firefox)",
			attr: "onafterscriptexecute",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onafterupdate event handler (IE)",
			attr: "onafterupdate",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onbeforeactivate event handler (IE)",
			attr: "onbeforeactivate",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onbeforedeactivate event handler (IE)",
			attr: "onbeforedeactivate",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onbeforeeditfocus event handler (IE)",
			attr: "onbeforeeditfocus",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onbeforescriptexecute event handler (Firefox)",
			attr: "onbeforescriptexecute",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onbeforeupdate event handler (IE)",
			attr: "onbeforeupdate",
			want: attributeTypeBlack,
		},
		{
			name: "Test with oncellchange event handler (IE)",
			attr: "oncellchange",
			want: attributeTypeBlack,
		},
		{
			name: "Test with ondatasetchanged event handler (IE)",
			attr: "ondatasetchanged",
			want: attributeTypeBlack,
		},
		{
			name: "Test with ondatasetcomplete event handler (IE)",
			attr: "ondatasetcomplete",
			want: attributeTypeBlack,
		},
		{
			name: "Test with ondeactivate event handler (IE)",
			attr: "ondeactivate",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onerrorupdate event handler (IE)",
			attr: "onerrorupdate",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onfilterchange event handler (IE)",
			attr: "onfilterchange",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onlayoutcomplete event handler (IE)",
			attr: "onlayoutcomplete",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onlosecapture event handler (IE)",
			attr: "onlosecapture",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onmozfullscreenchange event handler (Firefox)",
			attr: "onmozfullscreenchange",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onmozfullscreenerror event handler (Firefox)",
			attr: "onmozfullscreenerror",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onmozpointerlockchange event handler (Firefox)",
			attr: "onmozpointerlockchange",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onmozpointerlockerror event handler (Firefox)",
			attr: "onmozpointerlockerror",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onmsfullscreenchange event handler (IE/Edge)",
			attr: "onmsfullscreenchange",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onmsfullscreenerror event handler (IE/Edge)",
			attr: "onmsfullscreenerror",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onpropertychange event handler (IE)",
			attr: "onpropertychange",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onresizeend event handler (IE)",
			attr: "onresizeend",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onresizestart event handler (IE)",
			attr: "onresizestart",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onrowenter event handler (IE)",
			attr: "onrowenter",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onrowexit event handler (IE)",
			attr: "onrowexit",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onrowsdelete event handler (IE)",
			attr: "onrowsdelete",
			want: attributeTypeBlack,
		},
		{
			name: "Test with onrowsinserted event handler (IE)",
			attr: "onrowsinserted",
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
