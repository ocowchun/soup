package evaluator

import (
	"strings"
	"testing"

	"github.com/ocowchun/soup/lexer"
	"github.com/ocowchun/soup/parser"
)

func TestEvaluator_Builtin_Append(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{`(append '(1 2) '(3 4))`, `(1 2 3 4)`},
		{`(append '(1 2) '(3 4) '(5 6))`, `(1 2 3 4 5 6)`},
		{`(append '((1 2) (3 4)) '((5 6) (7 8)))`, `((1 2) (3 4) (5 6) (7 8))`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("expected %s, got %s", tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Builtin_ConOperation(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(car (cons 1 2))", `1`},
		{"(cdr (cons 1 2))", `2`},
		{"(car '(1 2 3))", `1`},
		{"(cdr '(1 2 3))", `(2 3)`},
		{"(car '((1 2) (3 4)))", `(1 2)`},
		{"(cdr '((1 2) (3 4)))", `((3 4))`},
		{"(cadr '(1 2 3))", `2`},
		{"(cdar '((1 2) (3 4)))", `(2)`},
		{"(caddr '((1 2) (3 4) (5 6)))", `(5 6)`},
		{"(cadddr '((1 2) (3 4) (5 6) (7 8)))", `(7 8)`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("expected %s, got %s", tt.expectedOutput, ret.String())
		}
	}
}

func testEval(input string, t *testing.T) ReturnValue {
	l := lexer.New(strings.NewReader(input))
	p := parser.New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	evaluator := New()
	result, err := evaluator.Eval(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return result
}
