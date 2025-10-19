package sema

import (
	"testing"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
)

func TestCheckLetStatement(t *testing.T) {
	input := `let x = 10
let y: int = 20
let z = "hello"`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestCheckConstReassignment(t *testing.T) {
	input := `const PI = 3.14
PI = 3.15`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected error for const reassignment")
	}

	// Error mesajı const assignment içermeli
	found := false
	for _, err := range errors {
		if semErr, ok := err.(*SemanticError); ok {
			if contains(semErr.Message, "const") {
				found = true
				break
			}
		}
	}

	if !found {
		t.Fatalf("expected const reassignment error, got: %v", errors)
	}
}

func TestCheckTypeMismatch(t *testing.T) {
	input := `let x: int = "hello"`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected type mismatch error")
	}
}

func TestCheckUndefinedVariable(t *testing.T) {
	input := `let x = y + 10`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected undefined variable error")
	}
}

func TestCheckFunctionReturnType(t *testing.T) {
	input := `function add(a: int, b: int): int
  return a + b
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestCheckReturnTypeMismatch(t *testing.T) {
	input := `function getValue(): int
  return "hello"
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected return type mismatch error")
	}
}

func TestCheckReturnOutsideFunction(t *testing.T) {
	input := `return 42`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected 'return outside function' error")
	}
}

func TestCheckFunctionScope(t *testing.T) {
	input := `function outer: int
  let x = 10
  function inner: int
    let y = x + 5
    return y
  end
  return inner()
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) != 0 {
		t.Fatalf("expected no errors for nested scopes, got: %v", errors)
	}
}

func TestCheckIfConditionType(t *testing.T) {
	input := `if 42
  print("error")
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected error: if condition must be bool")
	}
}

func TestCheckWhileConditionType(t *testing.T) {
	input := `while "hello"
  print("error")
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected error: while condition must be bool")
	}
}

func TestCheckBinaryOperatorTypes(t *testing.T) {
	tests := []struct {
		input       string
		shouldError bool
		desc        string
	}{
		{`let x = 10 + 20`, false, "int + int"},
		{`let x = 3.14 + 2.71`, false, "float + float"},
		{`let x = "hello" + "world"`, false, "string + string"},
		{`let x = 10 > 5`, false, "int > int"},
		{`let x = true && false`, false, "bool && bool"},
	}

	for _, tt := range tests {
		program := parseProgram(t, tt.input)
		checker := NewChecker()
		errors := checker.Check(program)

		if tt.shouldError && len(errors) == 0 {
			t.Errorf("%s: expected error but got none", tt.desc)
		}
		if !tt.shouldError && len(errors) != 0 {
			t.Errorf("%s: expected no error but got: %v", tt.desc, errors)
		}
	}
}

func TestCheckListTypes(t *testing.T) {
	input := `let numbers = [1, 2, 3]
let mixed = [1, "hello"]`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	// Mixed type list should produce error
	if len(errors) == 0 {
		t.Fatal("expected error for mixed type list")
	}
}

func TestCheckAwaitOutsideAsync(t *testing.T) {
	input := `function notAsync
  let x = await someFunc()
  return x
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected error: await outside async function")
	}
}

func TestCheckCallArgumentCount(t *testing.T) {
	input := `function add(a: int, b: int): int
  return a + b
end

function main
  let result = add(1)
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected error: wrong number of arguments")
	}
}

// Helper functions

func parseProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(input, "test.sky")
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	return program
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
