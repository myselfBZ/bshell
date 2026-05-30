package parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/myselfBZ/bshell/ast"
	"github.com/myselfBZ/bshell/lexer"
	"github.com/myselfBZ/bshell/token"
)

func TestParseSetEnv(t *testing.T) {
	tests := []struct{
		source string
		expected ast.SetVar
		err	 	 error
	}{
		{ source: "SET=KEY", expected: ast.SetVar{ Key: "SET", Val: "KEY",}, err: nil },
		{ source: "   SET=>  ", expected: ast.SetVar{}, err: fmt.Errorf("expected word got >") },
	}

	for _, test := range tests {
		l := lexer.New(test.source)
		p := New(l)

		cmds, err := p.Parse()

		if test.err == nil && err != nil {
			t.Fatalf("unexpected error. Input: %s. Error: %v", test.source, err)
		}

		if test.err != nil && err == nil {
			t.Fatalf("expected error: %v. got nil", test.err)
		}

		if test.err != nil && err != nil {
			if test.err.Error() != err.Error() {
				t.Fatalf("expected error %v got %v", test.err, err)
			}
		} else {



			cmd, ok := cmds[0].(*ast.SetVar)

			if !ok {
				t.Fatalf("expected type: ast.SetVar. got %T", cmds[0])
			}

			if cmd.Key != test.expected.Key {
				t.Fatalf("expected key: %s. got %s", test.expected.Key, cmd.Key)
			}

			if cmd.Val != test.expected.Val {
				t.Fatalf("expected val: %s. got %s", test.expected.Val, cmd.Val)
			}
		}
	}
}



func TestSimpleCommand(t *testing.T) {
	input := "cmd -la input.txt < input2.txt > output.txt > output1.txt > output2.txt &"

	expectedCmd := &ast.SimpleCommand{
		Name: "cmd",
		Args: []string{"-la", "input.txt"},
		IsBackground: true,
		Redirects: []ast.Redirect{
			{Type: ast.RedirectToStdin, Target: "input2.txt"},
			{Type: ast.RedirectStdout, Target: "output.txt"},
			{Type: ast.RedirectStdout, Target: "output1.txt"},
			{Type: ast.RedirectStdout, Target: "output2.txt"},
		},
	}



	l := lexer.New(input)

	p := New(l)

	cmds, err := p.Parse()

	if err != nil {
		t.Fatalf("unexpected error. Input: %s. Error: %v", input, err)
	}

	cmd, ok := cmds[0].(*ast.SimpleCommand)
	if !ok {
		t.Fatalf("expected type: ast.SimpleCommand. got %T", cmds[0])
	}

	if cmd.Name != expectedCmd.Name {
		t.Fatalf("SimpleCommand.Name: got %q, want %q", cmd.Name, expectedCmd.Name)
	}

	if len(cmd.Args) != len(expectedCmd.Args) {
		t.Fatalf("SimpleCommand.Args length: got %d (%v), want %d (%v)", len(cmd.Args), cmd.Args, len(expectedCmd.Args), expectedCmd.Args)
	}
	for i, arg := range expectedCmd.Args {
		if cmd.Args[i] != arg {
			t.Fatalf("SimpleCommand.Args[%d]: got %q, want %q", i, cmd.Args[i], arg)
		}
	}

	if cmd.IsBackground != expectedCmd.IsBackground {
		t.Fatalf("SimpleCommand.IsBackground: got %v, want %v", cmd.IsBackground, expectedCmd.IsBackground)
	}

	if len(cmd.Redirects) != len(expectedCmd.Redirects) {
		t.Fatalf("SimpleCommand.Redirects length: got %d (%v), want %d (%v)", len(cmd.Redirects), cmd.Redirects, len(expectedCmd.Redirects), expectedCmd.Redirects)
	}
	for i, redir := range expectedCmd.Redirects {
		if cmd.Redirects[i].Type != redir.Type {
			t.Fatalf("SimpleCommand.Redirects[%d].Type: got %v, want %v", i, cmd.Redirects[i].Type, redir.Type)
		}
		if cmd.Redirects[i].Target != redir.Target {
			t.Fatalf("SimpleCommand.Redirects[%d].Target: got %q, want %q", i, cmd.Redirects[i].Target, redir.Target)
		}

	}

}



func TestParseInfixExpressions(t *testing.T) {
	// Added a third pipe: | grep "pattern"
	input := "cmd hello | cat > output.txt | cat < input.txt | grep pattern"
	
	// 1. The original first pipe level
	firstPipe := &ast.InfixExpressionCmd{
		Left: &ast.SimpleCommand{
			Name: "cmd",
			Args: []string{"hello"},
		},
		Token:    token.NewToken(token.PIPE, "|"),
		Operator: "|",
		Right: &ast.SimpleCommand{
			Name: "cat",
			Redirects: []ast.Redirect{
				{Type: ast.RedirectStdout, Target: "output.txt"},
			},
		},
	}

	// 2. The original second pipe level (now wraps firstPipe)
	secondPipe := &ast.InfixExpressionCmd{
		Left:     firstPipe,
		Token:    token.NewToken(token.PIPE, "|"),
		Operator: "|",
		Right: &ast.SimpleCommand{
			Name: "cat",
			Redirects: []ast.Redirect{
				{Type: ast.RedirectToStdin, Target: "input.txt"},
			},
		},
	}

	// 3. The new top-level structure representing the final pipeline state
	expected := &ast.InfixExpressionCmd{
		Left:     secondPipe,
		Token:    token.NewToken(token.PIPE, "|"),
		Operator: "|",
		Right: &ast.SimpleCommand{
			Name: "grep",
			Args: []string{"pattern"},
		},
	}

	// Initialize Lexer and Parser
	l := lexer.New(input)
	p := New(l)

	// Execute Parse
	cmds, err := p.Parse()
	
	// 1. Check for unexpected parsing errors
	if err != nil {
		t.Fatalf("Parse() returned an unexpected error: %v", err)
	}

	// 2. Ensure we received exactly one top-level command statement
	if len(cmds) != 1 {
		// Note: removed the unsafe fmt.Printf(cmds[1]) that would panic if len was 0 or 1
		t.Fatalf("expected len(cmds) to be 1, got %d. Commands: %+v", len(cmds), cmds)
	}

	// 3. Assert the dynamic type of the parsed command is what we expect
	infixCmd, ok := cmds[0].(*ast.InfixExpressionCmd)
	if !ok {
		t.Fatalf("expected cmds[0] to be *ast.InfixExpressionCmd, got %T", cmds[0])
	}

	// 4. Deeply compare the parsed AST against your expected structure
	if !reflect.DeepEqual(infixCmd, expected) {
		t.Errorf("AST mismatch.\nExpected: %+v\nGot: %+v", expected, infixCmd)
	}
}

