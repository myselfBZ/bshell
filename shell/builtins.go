package shell

import (
	"fmt"
	"io"
	"os"
)


type BuiltInCmd struct {
	Name string
	Args []string
	In  io.Reader
	Out io.Writer
	Err io.Writer
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



