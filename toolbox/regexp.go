package toolbox

import (
	"regexp"

	log "github.com/sirupsen/logrus"
)

// InDomain returns true if the candidate hostname == or is sub-ordinate to the given domain,
// e.g. if domain is "foo.com" then "foo.com" and "sub.foo.com" will match, but "barfoo.com" won't.
func InDomain(domain, candidate string) bool {
	return MatchString(`(?i)(?:^|\.)`+domain+`$`, candidate)
}

// MatchString returns true if the candidate matches the pattern. This is just a wrapper
// around the regular regexp.MatchString, but checks for compilation errors and panics
// if one happens.
func MatchString(pattern, candidate string) bool {
	result, err := regexp.MatchString(pattern, candidate)
	if err != nil {
		log.Fatalf("Invalid pattern: %s", pattern)
	}
	return result
}
