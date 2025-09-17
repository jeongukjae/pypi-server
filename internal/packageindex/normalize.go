package packageindex

import (
	"regexp"
	"strings"
)

var normalizer = regexp.MustCompile(`[-_.]+`)

// NormalizePackageName normalizes a package name as per PEP 503.
//
// https://peps.python.org/pep-0503/#normalized-names
func NormalizePackageName(name string) string {
	name = strings.ToLower(name)
	name = normalizer.ReplaceAllString(name, "-")
	return name
}
