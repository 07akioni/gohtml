package lexer

import (
	"unicode/utf8"
)

func clonePosition(p Position) Position {
	return Position{
		Column: p.Column,
		Line:   p.Line,
		Index:  p.Index,
	}
}

func (l *Lexer) updatePosition(endIndex int) {
	startIndex := l.position.Index
	i := startIndex
	for i < endIndex {
		r, width := utf8.DecodeRuneInString(l.input[i:])
		if r == '\n' {
			l.position.Line += 1
			l.position.Column = 0
		} else {
			l.position.Column += 1
		}
		i += width
	}
	l.position.Index = i
}
