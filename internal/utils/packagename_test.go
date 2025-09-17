package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Foo_Bar", "foo-bar"},
		{"foo.bar", "foo-bar"},
		{"foo-bar", "foo-bar"},
		{"FOO.BAR_baz", "foo-bar-baz"},
		{"foo--bar__baz..qux", "foo-bar-baz-qux"},
		{"simple", "simple"},
		{"MiXeD_CaSe-Name", "mixed-case-name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizePackageName(tt.input)

			assert.Equal(t, tt.expected, got)
		})
	}
}
