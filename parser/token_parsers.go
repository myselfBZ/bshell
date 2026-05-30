package parser

import (
	"github.com/myselfBZ/bshell/ast"
	"github.com/myselfBZ/bshell/token"
)


func (p *Parser) parseWord() (ast.Command, error) {
	if p.peekTokenIs(token.EQUAL) {
		return p.parseEnvSet()
	}

	return p.parseSimpleCommand()
}
