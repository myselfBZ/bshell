package lexer

import (
	"testing"

	"github.com/myselfBZ/bshell/token"
)

func TestNextToken(t *testing.T) {
	tests := []struct {
		source   string
		expected token.Token
	}{
		{source: "       T       ", expected: token.NewToken(token.WORD, "T")},
		{source: "   ||  ", expected: token.NewToken(token.OR, "||")},
		{source: "|    ", expected: token.NewToken(token.PIPE, "|")},
		{source: " &&", expected: token.NewToken(token.AND, "&&")},
		{source: "&", expected: token.NewToken(token.AMPERSAND, "&")},
		{source: "&>", expected: token.NewToken(token.AMPERSAND_GT, "&>")},
		{source: ">>", expected: token.NewToken(token.GTGT, ">>")},
		{source: ">", expected: token.NewToken(token.GT, ">")},
		{source: " < ", expected: token.NewToken(token.LT, "<")},
		{source: " ; ", expected: token.NewToken(token.SEMICOLON, ";")},
		{source: " ( ", expected: token.NewToken(token.LEFT_PAR, "(")},
		{source: " ) ", expected: token.NewToken(token.RIGHT_PAR, ")")},
		{source: "$", expected: token.NewToken(token.DOLLAR_SIGN, "$")},
		{source: "=", expected: token.NewToken(token.EQUAL, "=")},
		{source: " 'cmd word word'  ", expected: token.NewToken(token.WORD, "cmd word word")},
		{source: " 'cmd' ", expected: token.NewToken(token.WORD, "cmd")},
		{source: " 'cmd ", expected: token.NewToken(token.WORD, "cmd ")},
		{source: " cmd'", expected: token.NewToken(token.WORD, "cmd")},
		{source: "\"cmd input.txt\"", expected: token.NewToken(token.WORD, "cmd input.txt")},
		{source: "\"cmd 'input.txt'\"", expected: token.NewToken(token.WORD, "cmd 'input.txt'")},
		{source: "2>", expected: token.NewToken(token.TWO_GT, "2>")},
	}

	for _, test := range tests {
		lexer := New(test.source)
		token := lexer.NextToken()

		if token.Type != test.expected.Type || token.Literal != test.expected.Literal {
			t.Fatalf(
				"expected token type %s token literal %s, got token type %s token literal %s. Lexer state: %v",
				test.expected.Type,
				test.expected.Literal,
				token.Type,
				token.Literal,
				lexer,
			)
		}
	}
}
