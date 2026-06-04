package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"sync"

	"github.com/myselfBZ/bshell/ast"
)

func New() *Shell {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return &Shell{
		cwd:  cwd,
		user: user,
		builtIns: map[string]bool{
			"cd":  true,
			"pwd": true,
			"type":true,
			"echo":true,
		},
	}
}

type Shell struct {
	cwd      string
	user     *user.User
	builtIns map[string]bool
}

func (s *Shell) Eval(cmds ...ast.Command) error {
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
				} else {
					left := node.Left.(*ast.InfixExpressionCmd)
					s.executePipe(left, pw)
				}
				pw.Close()

				wg.Wait()
			case "&&", "||":
				err := s.Eval(node.Left)
				if err != nil && node.Operator == "&&" {
					return err
				} else if err == nil && node.Operator == "&&" {
					right := node.Right.(*ast.SimpleCommand)
					return s.execSimpleCommand(right, os.Stdout, os.Stderr, os.Stdin)
				}

				if err != nil && node.Operator == "||" {
					right := node.Right.(*ast.SimpleCommand)
					return s.execSimpleCommand(right, os.Stdout, os.Stderr, os.Stdin)
				}

			}

		}
	}
	return nil
}

func (s *Shell) executePipe(node *ast.InfixExpressionCmd, out io.Writer) error {
	pr, pw, err := os.Pipe()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}

	right := node.Right.(*ast.SimpleCommand)

	go func() {
		s.execSimpleCommand(right, out, os.Stderr, pr)
		pr.Close()
	}()

	cmd, ok := node.Left.(*ast.SimpleCommand)

	if ok {
		err := s.execSimpleCommand(cmd, pw, os.Stderr, os.Stdin)
		pw.Close()

		if err != nil {
			return err
		}

		return nil
	}

	left := node.Left.(*ast.InfixExpressionCmd)

	err = s.executePipe(left, pw)
	return err
}

func (s *Shell) execSimpleCommand(cmd *ast.SimpleCommand, out io.Writer, stderr io.Writer, in io.Reader) error {
	for _, r := range cmd.Redirects {
		switch r.Type {
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

	if s.isBuiltIn(cmd.Name) {
		err := s.handleBuiltIn(&BuiltInCmd{
			Name: cmd.Name,
			Args: cmd.Args,
			In:   in,
			Out:  out,
			Err:  stderr,
		})

		if err != nil {
			fmt.Fprintln(stderr, err)
		}

		return err
	}

	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stdout = out
	c.Stderr = stderr
	c.Stdin = in

	return c.Run()
}

func (s *Shell) handleBuiltIn(cmd *BuiltInCmd) error {
	switch cmd.Name {
	case "cd":
		return s.cd(cmd.Args...)
	case "pwd":
		return s.pwd(cmd.Out)
	case "echo":
		return s.echo(cmd.Out, cmd.Args...)
	case "type":
		return s.typeCmd(cmd.Out, cmd.Args...)
	default:
		return fmt.Errorf("%s: command not found", cmd.Name)
	}
}

func (s *Shell) isBuiltIn(name string) bool {
	_, ok := s.builtIns[name]
	return ok
}
