package token

// valid tokens
const (
	WORD = "WORD"

	LEFT_PAR  = "("
	RIGHT_PAR = ")"

	// TBD...
	SINGLE_QUOTE = "\""
	DOUBLE_QUOTE = "'"

	LT           = "<"
	GT           = ">"
	TWO_GTGT     = "2>>"
	ONE_GTGT     = "1>>"


	TWO_GT       = "2>"
	AMPERSAND_GT = "&>"
	GTGT         = ">>"

	AMPERSAND   = "&"
	AND         = "&&"
	OR          = "||"
	SEMICOLON   = ";"
	PIPE        = "|"
	DOLLAR_SIGN = "$"
	EQUAL       = "="

	EOF = "EOF"
)

func NewToken(kind string, literal string) Token {
	return Token{Type: kind, Literal: literal}
}

type Token struct {
	Type    string
	Literal string
}
