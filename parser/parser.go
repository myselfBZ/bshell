package parser

import (
	"fmt"

	"github.com/myselfBZ/bshell/ast"
	"github.com/myselfBZ/bshell/lexer"
	"github.com/myselfBZ/bshell/token"
)

const (
	START = iota

	LOWEST

	AND_OR

	PIPE
)

var precedences = map[string]int{
	token.EOF: START,
	token.PIPE: PIPE,
	token.AND:  AND_OR,
	token.OR:   AND_OR,
}

func New(lexer *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:        lexer,
		infixFns:     make(map[string]InfixFn),
		tokenParsers: make(map[string]func() (ast.Command, error)),
	}
	p.curToken = p.lexer.NextToken()
	p.peekToken = p.lexer.NextToken()

	p.registerTokenParser(token.WORD, p.parseWord)
	p.registerInfixExpression(token.PIPE, p.parseInfixExpression)
	p.registerInfixExpression(token.OR, p.parseInfixExpression)
	p.registerInfixExpression(token.AND, p.parseInfixExpression)
	return p
}

type InfixFn func(left ast.Command) (ast.Command, error)

type Parser struct {
	lexer        *lexer.Lexer
	curToken     token.Token
	peekToken    token.Token
	tokenParsers map[string]func() (ast.Command, error)

	infixFns map[string]InfixFn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) registerTokenParser(key string, fn func() (ast.Command, error)) {
	p.tokenParsers[key] = fn
}

func (p *Parser) registerInfixExpression(key string, fn InfixFn) {
	p.infixFns[key] = fn
}

func (p *Parser) currentTokenIs(kind string) bool {
	return p.curToken.Type == kind
}

func (p *Parser) peekTokenIs(kind string) bool {
	return p.peekToken.Type == kind
}

func isInfixExpressionToken(kind string) bool {
	switch kind {
	case token.AND,
		token.OR,
		token.PIPE:
		return true
	}
	return false
}

func (p *Parser) Parse() ([]ast.Command, error) {
	cmds := []ast.Command{}

	for !p.currentTokenIs(token.EOF) {

		f, ok := p.tokenParsers[p.curToken.Type]
		if !ok {
			return nil, fmt.Errorf("invalid position for %s", p.curToken.Literal)
		}

		cmd, err := f()

		if err != nil {
			return nil, err
		}

		if isInfixExpressionToken(p.peekToken.Type) {
			cmd, err = p.parseExpression(cmd, START)
			if err != nil {
				return nil, fmt.Errorf("parse error: %v", err)
			}
			cmds = append(cmds, cmd)
		} else {
			cmds = append(cmds, cmd)
		}
		
		p.nextToken()

	}

	return cmds, nil
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) currentPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseRedirect(cmd *ast.SimpleCommand) error {
	rType := ast.Redirects[p.curToken.Literal]
	r := ast.Redirect{
		Type: rType,
	}

	if !p.peekTokenIs(token.WORD) {
		return fmt.Errorf("expected word for a redirect, got %s", p.curToken.Literal)
	}

	p.nextToken()
	r.Target = p.curToken.Literal

	cmd.Redirects = append(cmd.Redirects, r)
	return nil
}

func (p *Parser) parseInfixExpression(left ast.Command) (ast.Command, error) {
	node := &ast.InfixExpressionCmd{
		Left:     left,
		Operator: p.curToken.Literal,
		Token:    p.curToken,
	}
	c := p.currentPrecedence()
	p.nextToken()

	parseFn, ok := p.tokenParsers[p.curToken.Type]
	if !ok {
		return nil, fmt.Errorf("invalid position for %s", p.curToken.Literal)
	}

	cmd, err := parseFn()

	if err != nil {
		return nil, err
	}
	right, err := p.parseExpression(cmd, c)
	if err != nil {
		return nil, err
	}
	node.Right = right
	return node, nil
}

func (p *Parser) parseExpression(left ast.Command, precedence int) (ast.Command, error) {
	for precedence < p.peekPrecedence() {
		infixFn := p.infixFns[p.peekToken.Type]
		var err error
		p.nextToken()
		left, err = infixFn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parseCmdArgs(cmd *ast.SimpleCommand) error {
	cmd.Args = append(cmd.Args, p.curToken.Literal)
	return nil
}

func (p *Parser) parseSimpleCommand() (ast.Command, error) {
	simpleCommandTokens := map[string]func(cmd *ast.SimpleCommand) error{
		token.WORD:         p.parseCmdArgs,
		token.LT:           p.parseRedirect,
		token.GT:           p.parseRedirect,
		token.GTGT:         p.parseRedirect,
		token.TWO_GT:       p.parseRedirect,
		token.AMPERSAND_GT: p.parseRedirect,
	}
	cmd := &ast.SimpleCommand{Name: p.curToken.Literal}
	fn, ok := simpleCommandTokens[p.peekToken.Type]
	for ok {
		p.nextToken()

		err := fn(cmd)

		if err != nil {
			return nil, err
		}

		fn, ok = simpleCommandTokens[p.peekToken.Type]
	}

	if p.peekTokenIs(token.AMPERSAND) {
		p.nextToken()
		cmd.IsBackground = true

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	} else if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	} 

	return cmd, nil
}



func (p *Parser) parseEnvSet() (ast.Command, error) {
	cmd := &ast.SetVar{}

	if !p.currentTokenIs(token.WORD) {
		return nil, fmt.Errorf("expected word got %s", p.curToken.Literal)
	}

	cmd.Key = p.curToken.Literal

	p.nextToken()

	if !p.currentTokenIs(token.EQUAL) {
		return nil, fmt.Errorf("expected '=' got %s", p.curToken.Literal)
	}

	p.nextToken()

	if !p.currentTokenIs(token.WORD) {
		return nil, fmt.Errorf("expected word got %s", p.curToken.Literal)
	}

	cmd.Val = p.curToken.Literal
	return cmd, nil
}
