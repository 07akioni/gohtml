package lexer

import (
	"unicode/utf8"
)

func clonePosition(p Position) Position {
	return Position{
		column: p.column,
		line:   p.line,
		index:  p.index,
	}
}

func (l *Lexer) updatePosition(endIndex int) {
	startIndex := l.position.index
	i := startIndex
	for i < endIndex {
		r, width := utf8.DecodeRuneInString(l.input[i:])
		if r == '\n' {
			l.position.line += 1
			l.position.column = 0
		} else {
			l.position.column += 1
		}
		i += width
	}
	l.position.index = i
}
