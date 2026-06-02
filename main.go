package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/myselfBZ/bshell/ast"
	"github.com/myselfBZ/bshell/lexer"
	"github.com/myselfBZ/bshell/parser"
	"github.com/myselfBZ/bshell/shell"
)

const spaceSlope = 4

func walk(cmd ast.Command, spaces int) {
	if cmd == nil {
		return
	}

	spacesStr := strings.Repeat(" ", spaces)
	childSpaces := spaces + spaceSlope
	childSpacesStr := strings.Repeat(" ", childSpaces)

	switch c := cmd.(type) {
	case *ast.SimpleCommand:
		fmt.Printf("%s%s {\n", spacesStr, c.Name)
		fmt.Printf("%s    Args: %v\n", spacesStr, c.Args)
		fmt.Printf("%s    IsBackground: %v\n", spacesStr, c.IsBackground)

		for _, v := range c.Redirects {
			fmt.Printf("%s    Redirect: %+v\n", spacesStr, v)
		}
		fmt.Printf("%s}\n", spacesStr)

	case *ast.InfixExpressionCmd:
		fmt.Printf("%sInfixExpression (%s) {\n", spacesStr, c.Operator)
		
		fmt.Printf("%sLeft:\n", childSpacesStr)
		walk(c.Left, childSpaces+spaceSlope)

		fmt.Printf("%sRight:\n", childSpacesStr)
		walk(c.Right, childSpaces+spaceSlope)

		fmt.Printf("%s}\n", spacesStr)

	case *ast.GetVar:
		fmt.Printf("%sGetVar: $%s\n", spacesStr, c.Key)

	case *ast.SetVar:
		fmt.Printf("%sSetVar: %s=%s\n", spacesStr, c.Key, c.Val)
	}
}

func WalkPrint(cmd []ast.Command) {

	for _, c := range cmd{
		walk(c, 0)
	}
}

func main() {
	sh := shell.New()
	for {
		fmt.Print(">> ")
		r := bufio.NewReader(os.Stdin)
		input, _, err := r.ReadLine()

		if err != nil {
			fmt.Println("error reading a the input:", err)
			continue
		}

		if string(input) == "" {
			continue
		}

		l := lexer.New(string(input))
		p := parser.New(l)

		cmds, err := p.Parse()
		if err != nil {
			fmt.Println("parse error:", err)
			continue
		}

		sh.Eval(cmds...)
	}
}
