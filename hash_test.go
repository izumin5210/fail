package fail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeH(t *testing.T) {
	cases := []struct {
		test              string
		left, right, want H
	}{
		{
			test:  "empty",
			left:  H{},
			right: H{},
			want:  H{},
		},
		{
			test:  "simple",
			left:  H{"foo": 1},
			right: H{"bar": "baz"},
			want:  H{"foo": 1, "bar": "baz"},
		},
		{
			test:  "overwrite",
			left:  H{"foo": 1, "bar": "baz"},
			right: H{"qux": true, "foo": "quux"},
			want:  H{"foo": "quux", "bar": "baz", "qux": true},
		},
	}

	for _, c := range cases {
		t.Run(c.test, func(t *testing.T) {
			got := c.left.Merge(c.right)
			assert.Equal(t, c.want, got)
		})
	}
}
