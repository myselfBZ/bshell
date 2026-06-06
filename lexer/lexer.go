package lexer

import (
	"fmt"
	"unicode"

	"github.com/myselfBZ/bshell/token"
)


func New(input string) *Lexer {
	l := &Lexer{
		input: input,
	}
	l.ch = l.input[0]

	return l
}

type Lexer struct {
	ch            byte
	pos           int
	input         string
}

func (l *Lexer) skipWhiteSpace() {
	for l.ch == ' ' {
		l.readChar()
	}
}

func (l *Lexer) peek() byte {
	if l.pos <= len(l.input)-2 {
		return l.input[l.pos+1]
	}

	return 0
}

// do not mind the name
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || 
	unicode.IsDigit(rune(ch)) || 
	ch == '-' || 
	ch == '.' || 
	ch == '/' ||
	ch == '~' ||
	ch == '_' ||
	ch == ':'
}




func (l *Lexer) readWord() string {
	word := ""

	for isLetter(l.ch) || l.ch == '\\' {
		if l.ch == '\\' {
			l.readChar()
			word += string(l.ch)
		} else {
			word += string(l.ch)
		}
		l.readChar()
	}

	return word
}

func (l *Lexer) readChar() {
	if l.pos <= len(l.input)-1 {

		if l.pos == len(l.input)-1 {
			l.ch = 0
			return
		}

		l.pos++
		l.ch = l.input[l.pos]
	} else {
		l.ch = 0
	}
}

func (l *Lexer) readAdjacentContent() string {
	hasAdjacentContent := true
	content := ""
	for hasAdjacentContent {

		if l.ch == '\'' {
			quoteContent := l.readSingleQuote()
			content += quoteContent
			l.readChar()
		}

		if l.ch == '"' {
			quoteContent := l.readDoubleQuote()
			content += quoteContent
			l.readChar()
		}
	

		if isLetter(l.ch) {
			wordContent := l.readWord()
			content += wordContent
		}
		

		hasAdjacentContent = isLetter(l.ch) || l.ch == '\'' || l.ch == '"'
	}

	return content
}


func (l *Lexer) readDoubleQuote() string {
	l.readChar()

	word := ""
	for l.ch != '"' && l.ch != 0 {
		// escape sequence
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case '\\':
				word += "\\"
			case '"':
				word += "\""
			default:
				word += "\\" + string(l.ch)
			}
		} else {
			word += string(l.ch)
		}
		l.readChar()
	}
	return word
}

func (l *Lexer) readSingleQuote() string {
	l.readChar()
	word := ""
	for l.ch != '\'' && l.ch != 0 {
		word += string(l.ch)
		l.readChar()
	}
	return word
}

//
// func (l *Lexer) readQuote() string {
// 	l.readChar()
//
// 	word := ""
// 	for (l.inDoubleQuote || l.inSingleQuote) && l.ch != 0 {
//
// 		if l.ch == '\'' && l.inSingleQuote {
// 			l.inSingleQuote = false
// 			return word
// 		}
//
// 		if l.ch == '"' && l.inDoubleQuote {
// 			l.inDoubleQuote = false
// 			return word
// 		}
//
//
// 		word += string(l.ch)
// 		l.readChar()
// 	}
//
// 	return word
// }

func (l *Lexer) NextToken() token.Token {
	l.skipWhiteSpace()

	var t token.Token
	switch l.ch {
	case '\'':
		// l.inSingleQuote = true
		word := l.readSingleQuote()
		l.readChar()
		// handling adjacent words and quotes
		if isLetter(l.ch) || l.ch == '\'' || l.ch == '"'{
			content := l.readAdjacentContent()
			word += content
		}
		t = token.NewToken(token.WORD, word)
		return t
	case '"':
		// l.inDoubleQuote = true
		word := l.readDoubleQuote()
		l.readChar()
		if isLetter(l.ch) || l.ch == '\'' || l.ch == '"'{
			content := l.readAdjacentContent()
			word += content
		}
		t = token.NewToken(token.WORD, word)
	case ')':
		t = token.NewToken(token.RIGHT_PAR, string(l.ch))
	case '=':
		t = token.NewToken(token.EQUAL, string(l.ch))
	case '$':
		t = token.NewToken(token.DOLLAR_SIGN, string(l.ch))
	case '(':
		t = token.NewToken(token.LEFT_PAR, string(l.ch))
	case '<':
		t = token.NewToken(token.LT, string(l.ch))
	case ';':
		t = token.NewToken(token.SEMICOLON, string(l.ch))
	case '|':
		if l.peek() == '|' {
			t = token.NewToken(token.OR, "||")
			l.readChar()
		} else {
			t = token.NewToken(token.PIPE, string(l.ch))
		}
	case '>':
		if l.peek() == '>' {
			t = token.NewToken(token.GTGT, ">>")
			l.readChar()
		} else {
			t = token.NewToken(token.GT, string(l.ch))
		}
	case '&':
		if l.peek() == '&' {
			t = token.NewToken(token.AND, "&&")
			l.readChar()
		} else if l.peek() == '>' {
			t = token.NewToken(token.AMPERSAND_GT, "&>")
			l.readChar()
		} else {
			t = token.NewToken(token.AMPERSAND, string(l.ch))
		}
	case 0:
		t = token.NewToken(token.EOF, "EOF")
	default:
		if l.ch == '2' && l.peek() == '>'{
			l.readChar()
			// handling 2>>
			if l.peek() == '>' {
				l.readChar()
				t = token.NewToken(token.TWO_GTGT, "2>>")
			} else {
				t = token.NewToken(token.TWO_GT, "2>")
			}

		} else if l.ch == '1' && l.peek() == '>' {
			l.readChar()

			// handling 1>>
			if l.peek() == '>' {
				l.readChar()
				t = token.NewToken(token.ONE_GTGT, "1>>")
			} else {
				t = token.NewToken(token.GT, "1>")
			}

		} else {
			word := l.readWord()
			if word != "" {
				// adjacent quotes
				if isLetter(l.ch) || l.ch == '\'' || l.ch == '"'{
					content := l.readAdjacentContent()
					word += content
				}
				t = token.NewToken(token.WORD, word)

				return t
			} else {
				// TODO couldn't think of sth else
				fmt.Println("You're wrong current token:", string(l.ch))
				panic("readWord(): empty string")
			}
		}

	}

	l.readChar()
	return t
}

