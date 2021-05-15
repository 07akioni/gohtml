package lexer

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type Lexer struct {
	input    string
	position Position
	tokens   []Token
}

type Token struct {
	TokenType string        `json:"type"` // text, comment, tag-start(<), tag-end(>)
	Name      string        `json:"name"`
	Content   string        `json:"content"`
	Close     bool          `json:"close"`
	Position  TokenPosition `json:"position"`
}

type TokenPosition struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Index  int `json:"index"`
	Column int `json:"column"`
	Line   int `json:"line"`
}

func (l *Lexer) Lex() []Token {
	inputLength := len(l.input)
	for l.position.Index < inputLength {
		startIndex := l.position.Index
		l.lexText()
		if startIndex == l.position.Index {
			if strings.HasPrefix(l.input[startIndex:], "<!--") {
				l.lexComment()
			} else {
				l.lexTag(true)
			}
		}
	}
	return l.tokens
}

var alphanumeric = regexp.MustCompile("[A-Za-z0-9]")

func (l *Lexer) lexText() {
	startIndex := l.position.Index
	endIndex := findTextEnd(l.input, startIndex)
	startPosition := clonePosition(l.position)
	if endIndex == startIndex {
		return
	}
	l.updatePosition(endIndex)
	endPosition := clonePosition(l.position)
	l.tokens = append(l.tokens, Token{
		TokenType: "text",
		Content:   string(l.input[startIndex:endIndex]),
		Position: TokenPosition{
			Start: startPosition,
			End:   endPosition,
		},
	})
}

func findTextEnd(input string, index int) int {
	i := index
	for i < len(input) {
		r, width := utf8.DecodeRuneInString(input[i:])
		if r == '<' {
			if i+1 == len(input) {
				return i + 1
			}
			next := input[i+1]
			if next == '/' || next == '!' || alphanumeric.MatchString(string(next)) {
				return i
			}
		}
		i += width
	}
	return len(input)
}

func (l *Lexer) lexTag(pushToken bool) string {
	// Lex tag start `<`
	i := l.position.Index
	_, width := utf8.DecodeRuneInString(l.input[i:])
	nextRune, _ := utf8.DecodeRuneInString(l.input[i+width:])
	close := nextRune == '/'
	startPosition := clonePosition(l.position)
	if close {
		l.updatePosition(i + 2)
	} else {
		l.updatePosition(i + 1)
	}
	if pushToken {
		l.tokens = append(l.tokens, Token{
			TokenType: "tag-start",
			Close:     close,
			Position: TokenPosition{
				Start: startPosition,
			},
		})
	}
	// Lex tag name
	tagName := l.lexTagName(true)
	// Lex attrs
	l.lexTagAttrs()
	// Lex tag end `>`
	i = l.position.Index
	nextRune, _ = utf8.DecodeRuneInString(l.input[i:])
	close = nextRune == '/'
	if close {
		// for `/>`
		l.updatePosition(i + 2)
	} else {
		// for `>`
		l.updatePosition(i + 1)
	}
	endPosition := clonePosition(l.position)
	if pushToken {
		l.tokens = append(l.tokens, Token{
			TokenType: "tag-end",
			Close:     close,
			Position: TokenPosition{
				End: endPosition,
			},
		})
	}
	return tagName
}

// For cases:
// <  xxx  />
// <xxx  />
// <xxx>
// <xxx/>
// we need to find start index and end index
func (l *Lexer) lexTagName(pushToken bool) string {
	startIndex := l.position.Index
	for startIndex < len(l.input) {
		r, width := utf8.DecodeRuneInString(l.input[startIndex:])
		if !isWhiteSpace(r) && r != '>' && r != '/' {
			break
		}
		startIndex += width
	}
	l.updatePosition(startIndex)
	startPosition := clonePosition(l.position)
	endIndex := startIndex
	for endIndex < len(l.input) {
		r, width := utf8.DecodeRuneInString(l.input[endIndex:])
		if isWhiteSpace(r) || r == '>' || 'r' == '/' {
			break
		}
		endIndex += width
	}
	l.updatePosition(endIndex)

	tagName := l.input[startIndex:endIndex]

	endPosition := clonePosition(l.position)
	if pushToken {
		l.tokens = append(l.tokens, Token{
			TokenType: "tag",
			Content:   tagName,
			Position: TokenPosition{
				Start: startPosition,
				End:   endPosition,
			},
		})
	}
	return tagName
}

func (l *Lexer) lexTagAttrs() {
	quote := rune(0)
	// split words first
	// Basic: `key` `key=value` `key='value'` `key="value"`
	// Extra Case:
	//   - `='value'`
	//   - `="value"`
	//   - `=`
	//   - `'value'`
	//   - `"value"`
	//   - `key=`
	//   - For `key ='value'` `key = 'value'` `key= 'value'`
	//   - For `key ="value"` `key = "value"` `key= "value"`
	words := []string{}
	i := l.position.Index
	wordStartIndex := i
	for i < len(l.input) {
		r, width := utf8.DecodeRuneInString(l.input[i:])
		// quote
		if quote != 0 {
			i += width
			isQuoteEnd := (r == quote)
			if isQuoteEnd {
				quote = 0
			}
			continue
		}

		isTagEnd := r == '/' || r == '>'
		if isTagEnd {
			if i > wordStartIndex {
				words = append(words, l.input[wordStartIndex:i])
			}
			l.updatePosition(i)
			// do not consume '/' & '>'
			break
		}

		// <div**xxxx
		isWordEnd := isWhiteSpace(r)
		if isWordEnd {
			if wordStartIndex != i {
				words = append(words, l.input[wordStartIndex:i])
			}
			i += width
			wordStartIndex = i
			continue
		}

		isQuoteStart := r == '\'' || r == '"'
		if isQuoteStart {
			quote = r
		}
		i += width
	}

	for i = 0; i < len(words); i += 1 {
		word := words[i]
		if strings.HasSuffix(word, "=") {
			// `key=`, need value
			name := word[0 : len(word)-1]
			i += 1
			value := words[i]
			l.tokens = append(l.tokens, Token{
				TokenType: "attr",
				Name:      name,
				Content:   value,
			})
			continue
		}

		eqIndex := strings.IndexRune(word, '=')
		if eqIndex != -1 {
			// `key=value`, need no extra data
			name := word[0:eqIndex]
			value := word[eqIndex:]
			l.tokens = append(l.tokens, Token{
				TokenType: "attr",
				Name:      name,
				Content:   value,
			})
			continue
		}

		if i+1 >= len(words) {
			l.tokens = append(l.tokens, Token{
				TokenType: "attr",
				Name:      word,
			})
			break
		}

		nextWord := words[i+1]
		nextRune, _ := utf8.DecodeRuneInString(nextWord)
		if nextRune != '=' {
			l.tokens = append(l.tokens, Token{
				TokenType: "attr",
				Name:      word,
			})
		} else {
			if nextWord == "=" {
				l.tokens = append(l.tokens, Token{
					TokenType: "attr",
					Name:      word,
					Content:   words[i+2],
				})
				i += 2
			} else {
				l.tokens = append(l.tokens, Token{
					TokenType: "attr",
					Name:      word,
					Content:   words[i+1][1:],
				})
				i += 1
			}
		}
	}
}

func (l *Lexer) lexComment() {
	commentStartIndex := l.position.Index
	contentEndIndex := strings.Index(l.input[commentStartIndex:], "-->")
	commentEndIndex := contentEndIndex + 3
	if contentEndIndex == -1 {
		contentEndIndex = len(l.input)
		commentEndIndex = contentEndIndex
	}
	commentStartPosition := clonePosition(l.position)
	l.updatePosition(contentEndIndex)
	commentEndPosition := clonePosition(l.position)
	l.updatePosition(commentEndIndex)
	l.tokens = append(l.tokens, Token{
		TokenType: "comment",
		Content:   l.input[commentStartIndex:commentEndIndex],
		Position: TokenPosition{
			Start: commentStartPosition,
			End:   commentEndPosition,
		},
	})
}

// some tag need to be skipped, for example <script>xxx</script>
func (l *Lexer) lexSkipTag(tagName string) {
	i := l.position.Index
	startPosition := clonePosition(l.position) // text start position
	for i < len(l.input) {
		closeTagIndex := i + strings.Index(l.input[i:], "</"+tagName)
		if closeTagIndex == -1 {
			l.lexText()
			break
		}
		l.updatePosition(closeTagIndex)
		tagStartPosition := clonePosition(l.position)
		if tagName == l.lexTag(false) {
			l.tokens = append(l.tokens, Token{
				TokenType: "text",
				Content:   l.input[startPosition.Index:l.position.Index],
				Position: TokenPosition{
					Start: startPosition,
					End:   clonePosition(l.position),
				},
			})
			l.position = tagStartPosition
			l.lexTag(true)
		}
	}
}

func MakeLexer(input string) *Lexer {
	return &Lexer{
		input: input,
	}
}
