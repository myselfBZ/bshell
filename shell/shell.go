package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/myselfBZ/bshell/ast"
)

func New() *Shell {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return &Shell{
		cwd: cwd,
	}
}

type Shell struct {
	cwd string
	builtInCmds map[string]func(args ...string) error
}


func (s *Shell) Eval(cmds []ast.Command) error {
	for _, c := range cmds {
		switch node := c.(type) {
		case *ast.SimpleCommand:
			return s.execSimpleCommand(node, os.Stdout, os.Stdin, os.Stderr) 
		case *ast.InfixExpressionCmd:

			switch node.Operator {
			case "|":
				var wg sync.WaitGroup
				pr, pw, err := os.Pipe()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					return err
				}

				right, ok := node.Right.(*ast.SimpleCommand)

				if !ok {
					fmt.Printf("how come the right node of pipe is not a command: %T\n", node.Right)
					panic("This cannot be")
				}

				wg.Go(func() {
					s.execSimpleCommand(right, os.Stdout, os.Stderr, pr)
					pr.Close()
				})


				left, ok := node.Left.(*ast.SimpleCommand)

				if ok {
					s.execSimpleCommand(left, pw, os.Stderr, os.Stdin)
					pw.Close()
				}

				wg.Wait()


			case "&&":
			case "||":
			}

		}
	}
	return nil
}





func (s *Shell) execSimpleCommand(cmd *ast.SimpleCommand, out io.Writer, stderr io.Writer, in io.Reader) error {
	for _, r := range cmd.Redirects {
		switch r.Type{
		case ast.RedirectStdout:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			out = f
			defer f.Close()
		case ast.RedirectStdoutAppend:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			out = f
			defer f.Close()
		case ast.RedirectStdErr:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			stderr = f
			defer f.Close()
		case ast.RedirectStdErrAndOut:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			out = f
			stderr = f
			defer f.Close()
		case ast.RedirectToStdin:
			f, err := os.Open(r.Target)
			if err != nil {
				return err
			}
			in = f
			defer f.Close()
		}
	}

	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stdout = out
	c.Stderr = stderr
	c.Stdin = in
	return c.Run()
}
