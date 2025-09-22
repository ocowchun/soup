package parser

import (
	"strings"
	"testing"

	"github.com/ocowchun/soup/lexer"
)

// TODO: add test cases for errors

func TestParser_ParseNumber(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"123", "123"},
		{"45.67", "45.67"},
		{"-89", "-89"},
		{"+9527", "+9527"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}
		numLiteral, ok := program.Expressions[0].(*NumberLiteral)
		if !ok {
			t.Fatalf("expected CallExpression, got %T", program.Expressions[0])
		}

		if numLiteral.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, numLiteral.String())
		}
	}
}

func TestParser_ParseFunctionExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(+ 1 2)", "(+ 1 2)"},
		{"(- 10 2)", "(- 10 2)"},
		{"(* 3 4)", "(* 3 4)"},
		{"(/ 10 2)", "(/ 10 2)"},
		{"(+ 1 (* 2 3))", "(+ 1 (* 2 3))"},
		{"(fib 5)", "(fib 5)"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}
		funcExpr, ok := program.Expressions[0].(*CallExpression)
		if !ok {
			t.Fatalf("expected CallExpression, got %T", program.Expressions[0])
		}

		if funcExpr.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, funcExpr.String())
		}
	}
}

func TestParser_ParseIfExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{`(if (> x 0) "positive" "non-positive")`, `(if (> x 0) "positive" "non-positive")`},
		{`(if (> x 0) x)`, `(if (> x 0) x )`},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}
		expr, ok := program.Expressions[0].(*IfExpression)
		if !ok {
			t.Fatalf("expected CallExpression, got %T", program.Expressions[0])
		}

		if expr.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, expr.String())
		}
	}
}

func TestParser_ParseDefineExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(define foo 123)", "(define foo 123)"},
		{"(define (add a b) (+ a b))", "(define (add a b) (+ a b))"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}
		defineExpr, ok := program.Expressions[0].(*DefineExpression)
		if !ok {
			t.Fatalf("expected DefineExpression, got %T", program.Expressions[0])
		}

		if defineExpr.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, defineExpr.String())
		}
	}
}

func TestParser_ParseLambdaExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(lambda (a b) (+ a b))", "(lambda (a b) (+ a b))"},
		{"(lambda () 123)", "(lambda () 123)"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}
		lambdaExpr, ok := program.Expressions[0].(*LambdaExpression)
		if !ok {
			t.Fatalf("expected DefineExpression, got %T", program.Expressions[0])
		}

		if lambdaExpr.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, lambdaExpr.String())
		}
	}
}

func TestParser_ParseLetExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(let ((a 1)) (+ a 5))", "((lambda (a) (+ a 5)) 1)"},
		{"(let ((a 1) (b 2)) (+ a b))", "((lambda (a b) (+ a b)) 1 2)"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}
		callExpr, ok := program.Expressions[0].(*CallExpression)
		if !ok {
			t.Fatalf("expected DefineExpression, got %T", program.Expressions[0])
		}

		if callExpr.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, callExpr.String())
		}
	}
}

func TestParser_ParseCondExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(cond ((> a 0) 123) (else 234))", "(if (> a 0) 123 234)"},
		{"(cond ((> a 0) 123) ((> b 0) 234) (else 456))", "(if (> a 0) 123 (if (> b 0) 234 456))"},
		{"(cond ((> a 0) 123))", "(if (> a 0) 123 )"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}
		ifExpression, ok := program.Expressions[0].(*IfExpression)
		if !ok {
			t.Fatalf("expected DefineExpression, got %T", program.Expressions[0])
		}

		if ifExpression.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, ifExpression.String())
		}
	}
}

func TestParser_ParseQuoteExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"'123", "123"},
		{"'foo", "'foo"},
		{"'define", "'define"},
		{"'(1 2 3)", "'(1 2 3)"},
		{"'\"hola\"", "\"hola\""},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}

		exp := program.Expressions[0]
		if exp.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, exp.String())
		}
	}
}

func TestParser_ParseSetExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(set! a 123)", "(set! a 123)"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}

		exp := program.Expressions[0]
		if exp.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, exp.String())
		}
	}
}

func TestParser_ParseDelayExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(delay (+ 1 2))", "(delay (+ 1 2))"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}

		exp := program.Expressions[0]
		if exp.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, exp.String())
		}
	}
}

func TestParser_ParseStreamExpression(t *testing.T) {
	tests := []struct {
		input          string
		expectedString string
	}{
		{"(cons-stream 1 2)", "(cons-stream 1 2)"},
	}
	for _, tt := range tests {
		text := tt.input
		l := lexer.New(strings.NewReader(text))
		p := New(l)

		program, err := p.Parse()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(program.Expressions) != 1 {
			t.Fatalf("expected 1 expression, got %d", len(program.Expressions))
		}

		exp := program.Expressions[0]
		if exp.String() != tt.expectedString {
			t.Fatalf("expected string representation '%s', got %s", tt.expectedString, exp.String())
		}
	}
}
