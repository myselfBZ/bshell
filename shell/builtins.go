package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)


type BuiltInCmd struct {
	Name string
	Args []string
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func (s *Shell) typeCmd(out io.Writer, args ...string) error {
	for _, c := range args {
		if _, ok := s.builtIns[c]; ok {
			fmt.Printf("%s is a shell builtin\n", c)
		} else if !ok {
			p, err := exec.LookPath(args[0])
			if err != nil {
				fmt.Printf("%s: not found\n", c)
			} else {
				fmt.Printf("%s is %s\n", c, p)
			}
		}
	}
	return nil
}

func (s *Shell) echo(out io.Writer, args ...string) error {
	for _, v := range args {
		_, err := fmt.Fprint(out, v)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(out)
	return err
}

func (s *Shell) cd(args ...string) error {
	targetDir := ""
	if len(args) == 0 {
		targetDir = s.user.HomeDir
	} else {
		targetDir = args[0]
	}
	err := os.Chdir(targetDir)
	if err != nil {
		return fmt.Errorf("cd: %v", err)
	}
	// absolute path
	curd, err := os.Getwd()

	if err != nil {
		return fmt.Errorf("cd: %v", err)
	}
	s.cwd = curd
	return nil
}

func (s *Shell) pwd(out io.Writer) error {
	_, err := fmt.Fprintln(out, s.cwd)
	return err
}



