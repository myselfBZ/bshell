package ast

import "github.com/myselfBZ/bshell/token"


type RedirectType int

const (
	_ RedirectType = iota

	RedirectStdout
	RedirectToStdin
	RedirectStdErr
	RedirectStdErrAndOut
	RedirectStdoutAppend
)

var Redirects = map[string]RedirectType{
	">":  RedirectStdout,
	"<":  RedirectToStdin,
	">>": RedirectStdoutAppend,
	"&>": RedirectStdErrAndOut,
	"2>": RedirectStdErr,
}

var (
	_ Command = (*GetVar)(nil)
	_ Command = (*SetVar)(nil)
	_ Command = (*SimpleCommand)(nil)
	_ Command = (*InfixExpressionCmd)(nil)
)

type Command interface {
	command()
}

type GetVar struct {
	Key string
}

func (g *GetVar) command() {}

type SetVar struct {
	Key string
	Val string
}

func (s *SetVar) command() {}

type SimpleCommand struct {
	Name         string
	Args         []string
	IsBackground bool
	Redirects    []Redirect
}

func (s *SimpleCommand) command() {}

type Redirect struct {
	Type   RedirectType
	Target string
}

type InfixExpressionCmd struct {
	Left     Command
	Right    Command
	Operator string

	Token token.Token
}

func (p *InfixExpressionCmd) command() {}
