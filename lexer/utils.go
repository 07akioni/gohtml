package lexer

import "regexp"

func minInt(a int, b int) int {
	if a > b {
		return b
	}
	return a
}

var whiteSpaceRe = regexp.MustCompile("\\s")

func isWhiteSpace(r rune) bool {
	return whiteSpaceRe.MatchString(string(r))
}
