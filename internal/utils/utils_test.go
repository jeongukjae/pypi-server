package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCheck(t *testing.T) {
	tests := []string{
		"1.2.3",
		"1.2.3a1",
		"2!1.2.3",
		"1.2.3.post4",
		"1.2.3.dev5",
		"1.2rc3",
		"1.2.3+abc",
		"1.2.3.post4.dev5+abc",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			v, err := ParseVersion(input)
			require.NoError(t, err)
			assert.Equal(t, input, v.String())
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input     string
		wantError bool
		want      *Version
	}{
		{
			"1.2.3",
			false,
			&Version{
				Releases: []int64{1, 2, 3},
			},
		},
		{
			"1.2.3a1",
			false,
			&Version{
				Releases:      []int64{1, 2, 3},
				PreReleaseTag: "a",
				PreRelease:    Pointer(int64(1)),
			},
		},
		{
			"2!1.2.3",
			false,
			&Version{
				Epoch:    Pointer(int64(2)),
				Releases: []int64{1, 2, 3},
			},
		},
		{
			"1.2.3.post4",
			false,
			&Version{
				Releases:    []int64{1, 2, 3},
				PostRelease: Pointer(int64(4)),
			},
		},
		{
			"1.2.3.dev5",
			false,
			&Version{
				Releases:   []int64{1, 2, 3},
				DevRelease: Pointer(int64(5)),
			},
		},
		{
			"1.2rc3",
			false,
			&Version{
				Releases:      []int64{1, 2},
				PreReleaseTag: "rc",
				PreRelease:    Pointer(int64(3)),
			},
		},
		{
			"1.2.3+abc",
			false,
			&Version{
				Releases: []int64{1, 2, 3},
				Local:    Pointer("abc"),
			},
		},
		{
			"1.2.3.post4.dev5+abc",
			false,
			&Version{
				Releases:    []int64{1, 2, 3},
				PostRelease: Pointer(int64(4)),
				DevRelease:  Pointer(int64(5)),
				Local:       Pointer("abc"),
			},
		},
		{
			"invalid",
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseVersion(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		v1, v2 string
		want   int
	}{
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.0", 0},
		{"1.0.0a1", "1.0.0", -1}, // pre-release < release
		{"1.0.0", "1.0.0a1", 1},
		{"1.0.0a1", "1.0.0a2", -1},
		{"1.0.0a2", "1.0.0a1", 1},
		{"1.0.0.post1", "1.0.0", 1}, // post-release > release
		{"1.0.0", "1.0.0.post1", -1},
		{"1.0.0.dev1", "1.0.0", -1}, // dev-release < release
		{"1.0.0", "1.0.0.dev1", 1},
		{"1.0.0+abc", "1.0.0+abd", -1}, // local segment lexicographical
		{"1.0.0+abd", "1.0.0+abc", 1},
		{"1.0.0+abc", "1.0.0+abc", 0},
		{"2!1.0.0", "1!1.0.0", 1}, // epoch
		{"1!1.0.0", "2!1.0.0", -1},
	}

	for _, tt := range tests {
		v1, err1 := ParseVersion(tt.v1)
		v2, err2 := ParseVersion(tt.v2)
		if err1 != nil || err2 != nil {
			t.Fatalf("failed to parse version: %v, %v", err1, err2)
		}
		got := v1.Compare(v2)
		if got != tt.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
		}
	}
}
