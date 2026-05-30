package shell

import (
	"os"
	"os/exec"

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
			return s.execSimpleCommand(node) 
		}
	}
	return nil
}



func (s *Shell) execSimpleCommand(cmd *ast.SimpleCommand) error {
	defaultOut := os.Stdout
	defaultErr := os.Stderr
	defaultIn := os.Stdin

	for _, r := range cmd.Redirects {
		switch r.Type{
		case ast.RedirectStdout:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defaultOut = f
		case ast.RedirectStdoutAppend:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defaultOut = f
		case ast.RedirectStdErr:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defaultErr = f
		case ast.RedirectStdErrAndOut:
			f, err := os.OpenFile(r.Target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defaultOut = f
			defaultErr = f
		case ast.RedirectToStdin:
			f, err := os.Open(r.Target)
			if err != nil {
				return err
			}
			defaultIn = f
		}
	}

	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stdout = defaultOut
	c.Stderr = defaultErr
	c.Stdin = defaultIn
	return c.Run()
}
