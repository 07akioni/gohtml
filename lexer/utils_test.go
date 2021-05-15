package lexer

import "testing"

func TestIsWhiteSpace(t *testing.T) {
	if isWhiteSpace(' ') != true {
		t.Error("' ' should be whitespace")
	}
	if isWhiteSpace('\n') != true {
		t.Error("'\\n' should be whitespace")
	}
	if isWhiteSpace('x') != false {
		t.Error("'x' should not be whitespace")
	}
}
