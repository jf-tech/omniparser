package idr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveLastFilterInXPath(t *testing.T) {
	for _, test := range []struct {
		name   string
		xpath  string
		expect string
	}{
		{
			name:   "empty",
			xpath:  "",
			expect: "",
		},
		{
			name:   "blank",
			xpath:  "    ",
			expect: "    ",
		},
		{
			name:   " /A/B/C ",
			xpath:  " /A/B/C ",
			expect: " /A/B/C ",
		},
		{
			name:   "unbalanced brackets",
			xpath:  "/A/B/C[...]]",
			expect: "/A/B/C[...]]",
		},
		{
			name:   "another unbalanced brackets",
			xpath:  "/A/B/C']",
			expect: "/A/B/C']",
		},
		{
			name:   "balanced brackets",
			xpath:  "/A/B/C[...]",
			expect: "/A/B/C",
		},
		{
			name:   "brackets in single quotes",
			xpath:  "/A/B/C[.='[']",
			expect: "/A/B/C",
		},
		{
			name:   "brackets in double quotes",
			xpath:  `/A/B/C[.="abc]"]`,
			expect: "/A/B/C",
		},
		{
			name:   "brackets not at the end",
			xpath:  `/A/B/C[.="abc]"]/D`,
			expect: `/A/B/C[.="abc]"]/D`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, removeLastFilterInXPath(test.xpath))
		})
	}
}
