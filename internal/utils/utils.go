package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Epoch         *int64
	Releases      []int64
	PreReleaseTag string
	PreRelease    *int64
	PostRelease   *int64
	DevRelease    *int64
	Local         *string
}

//nolint:gocyclo // Complexity is acceptable here for the sake of clarity.
func (v *Version) Compare(v2 *Version) int {
	// 1: v < other
	// 0: v == other
	// -1: v > other

	// 1. Epoch
	switch {
	case v.Epoch != nil && v2.Epoch != nil:
		if *v.Epoch > *v2.Epoch {
			return 1
		}
		if *v.Epoch < *v2.Epoch {
			return -1
		}
	case v.Epoch != nil:
		return 1
	case v2.Epoch != nil:
		return -1
	}

	// 2. Release
	maxLen := max(len(v2.Releases), len(v.Releases))

	for i := range maxLen {
		r1 := int64(0)
		if i < len(v.Releases) {
			r1 = v.Releases[i]
		}
		r2 := int64(0)
		if i < len(v2.Releases) {
			r2 = v2.Releases[i]
		}
		if r1 > r2 {
			return 1
		}
		if r1 < r2 {
			return -1
		}
	}

	// 3. Pre-release
	if v.PreRelease == nil && v2.PreRelease != nil {
		return 1
	}
	if v.PreRelease != nil && v2.PreRelease == nil {
		return -1
	}

	preReleaseOrder := map[string]int{"a": 1, "b": 2, "rc": 3}
	vPreOrder := preReleaseOrder[strings.ToLower(v.PreReleaseTag)]
	v2PreOrder := preReleaseOrder[strings.ToLower(v2.PreReleaseTag)]

	if vPreOrder > v2PreOrder {
		return 1
	}
	if vPreOrder < v2PreOrder {
		return -1
	}

	switch {
	case v.PreRelease != nil && v2.PreRelease != nil:
		if *v.PreRelease > *v2.PreRelease {
			return 1
		}
		if *v.PreRelease < *v2.PreRelease {
			return -1
		}
	case v.PreRelease != nil:
		// A release with a pre-release is always smaller than a full release.
		return -1
	case v2.PreRelease != nil:
		return 1
	}

	// 4. Post-release
	switch {
	case v.PostRelease != nil && v2.PostRelease != nil:
		if *v.PostRelease > *v2.PostRelease {
			return 1
		}
		if *v.PostRelease < *v2.PostRelease {
			return -1
		}
	case v.PostRelease != nil:
		return 1
	case v2.PostRelease != nil:
		return -1
	}

	// 5. Dev-release
	switch {
	case v.DevRelease != nil && v2.DevRelease != nil:
		if *v.DevRelease > *v2.DevRelease {
			return 1
		}
		if *v.DevRelease < *v2.DevRelease {
			return -1
		}
	case v.DevRelease != nil:
		return -1
	case v2.DevRelease != nil:
		return 1
	}

	// 6. Local version identifier
	switch {
	case v.Local != nil && v2.Local != nil:
		if *v.Local > *v2.Local {
			return 1
		}
		if *v.Local < *v2.Local {
			return -1
		}
	case v.Local != nil:
		return 1
	case v2.Local != nil:
		return -1
	}

	return 0
}

func (v *Version) String() string {
	var b strings.Builder

	if v.Epoch != nil {
		fmt.Fprintf(&b, "%d!", *v.Epoch)
	}

	for i, r := range v.Releases {
		if i > 0 {
			b.WriteString(".")
		}
		fmt.Fprintf(&b, "%d", r)
	}

	if v.PreReleaseTag != "" && v.PreRelease != nil {
		fmt.Fprintf(&b, "%s%d", v.PreReleaseTag, *v.PreRelease)
	}

	if v.PostRelease != nil {
		fmt.Fprintf(&b, ".post%d", *v.PostRelease)
	}

	if v.DevRelease != nil {
		fmt.Fprintf(&b, ".dev%d", *v.DevRelease)
	}

	if v.Local != nil {
		fmt.Fprintf(&b, "+%s", *v.Local)
	}

	return b.String()
}

func ParseVersion(version string) (*Version, error) {
	// Full version: [N!]N(.N)*[{a|b|rc}N][.postN][.devN]
	//
	// Epoch: ([1-9][0-9]*!)?
	// Release: (0|[1-9][0-9]*)(\.(0|[1-9][0-9]*))*
	// Pre-release: ((a|b|rc)(0|[1-9][0-9]*))?
	// Post-release: (\.post(0|[1-9][0-9]*))?
	// Developmental release: (\.dev(0|[1-9][0-9]*))?
	//
	// And.. Local version segment might be there:

	version = strings.TrimSpace(version)
	version = strings.ToLower(version)

	re := regexp.MustCompile(`^(([1-9][0-9]*)!)?(0|[1-9][0-9]*)((\.(0|[1-9][0-9]*))*)((a|b|rc|alpha|beta|c)(0|[1-9][0-9]*))?(\.post(0|[1-9][0-9]*))?(\.dev(0|[1-9][0-9]*))?(\+([a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*))?$`)
	matches := re.FindStringSubmatch(version)
	if matches == nil {
		return nil, fmt.Errorf("invalid version: %q", version)
	}

	v := &Version{}
	if matches[2] != "" {
		epoch, err := strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid epoch in version: %q", version)
		}

		v.Epoch = &epoch
	}

	releases := strings.Split(matches[3], ".")
	if matches[4] != "" {
		releases = append(releases, strings.Split(strings.TrimLeft(matches[4], "."), ".")...)
	}

	// normalize the number of releases
	for i := range releases {
		for len(releases[i]) > 1 && releases[i][0] == '0' {
			releases[i] = releases[i][1:]
		}

		if releases[i] == "" {
			return nil, fmt.Errorf("release segment must not be empty: %q", version)
		}

		segment, err := strconv.ParseInt(releases[i], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid release segment in version: %q", version)
		}

		v.Releases = append(v.Releases, segment)
	}

	if matches[7] != "" {
		v.PreReleaseTag = matches[8]
		if v.PreReleaseTag == "c" {
			v.PreReleaseTag = "rc"
		}
		if v.PreReleaseTag == "alpha" {
			v.PreReleaseTag = "a"
		}
		if v.PreReleaseTag == "beta" {
			v.PreReleaseTag = "b"
		}

		preRelease, err := strconv.ParseInt(matches[9], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid pre-release segment in version: %q", version)
		}
		v.PreRelease = &preRelease
	}

	if matches[10] != "" {
		postRelease, err := strconv.ParseInt(matches[11], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid post-release segment in version: %q", version)
		}
		v.PostRelease = &postRelease
	}

	if matches[12] != "" {
		devRelease, err := strconv.ParseInt(matches[13], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid dev-release segment in version: %q", version)
		}
		v.DevRelease = &devRelease
	}

	if matches[15] != "" {
		v.Local = &matches[15]
	}

	return v, nil
}
