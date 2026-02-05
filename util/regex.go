package util

import (
	"regexp"
	"strings"
)

func GlobToRegex(pattern string) string {

	var finalPattern strings.Builder

	finalPattern.WriteString("^")

	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]

		switch ch {
		case '*':
			finalPattern.WriteString(`.*`)
		case '?':
			finalPattern.WriteString(`.`)
		default:
			finalPattern.WriteString(regexp.QuoteMeta(string(ch)))
		}
	}

	finalPattern.WriteString("$")
	return finalPattern.String()

}

func MatchString(pattern string, val string) (bool, error) {
	finalPattern := GlobToRegex(pattern)
	return regexp.MatchString(finalPattern, val)
}
