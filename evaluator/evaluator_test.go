package evaluator

import (
	"strings"
	"testing"

	"github.com/ocowchun/soup/lexer"
	"github.com/ocowchun/soup/parser"
)

func TestEvaluator_Builtin_List(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{`(list 1 2 3)`, `(1 2 3)`},
		{`(list (list 4 5) (list 6))`, `((4 5) (6))`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

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
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Builtin_Map(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{`(map (lambda (x) (+ x 1)) '(1 2))`, `(2 3)`},
		{`(map + '(1 2))`, `(1 2)`},
		{`(map (lambda (x y) (+ x y)) '(1 2) '(3 4))`, `(4 6)`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Builtin_MathOperation(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(+ 1 2)", `3`},
		{"(+ 1 2 3)", `6`},
		{"(+ 1)", `1`},
		{"(- 5 2)", `3`},
		{"(- 1 3)", `-2`},
		{"(- 1 2 3)", `-4`},
		{"(- 1)", `-1`},
		{"(* 1)", `1`},
		{"(* 2)", `2`},
		{"(* 2 3)", `6`},
		{"(/ 1)", `1`},
		{"(/ 2)", `2`},
		{"(/ 5 10)", `0.5`},
		{"(/ 2 3)", `0.6666666666666666`},
		{"(remainder 2 3)", `2`},
		{"(remainder 12 3)", `0`},
		{"(remainder 5 3)", `2`},
		{"(sqrt 4)", `2`},
		{"(abs 4)", `4`},
		{"(abs -4)", `4`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Builtin_EqAndCompare(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(number? 1)", `#t`},
		{"(number? 'a)", `#f`},
		{"(symbol? 1)", `#f`},
		{"(symbol? 'a)", `#t`},
		{"(string? 1)", `#f`},
		{"(string? 'a)", `#f`},
		{"(string? \"foo\")", `#t`},
		{"(pair? 'a)", `#f`},
		{"(pair? '(1 2 3))", `#t`},
		{"(pair? '())", `#f`},
		{"(pair? (cons 1 2))", `#t`},
		{"(list? '(1 2 3))", `#t`},
		{"(list? '())", `#t`},
		{"(list? (cons 1 2))", `#f`},
		{"(eq? 'a 'a)", `#t`},
		{"(eq? 1 1)", `#t`},
		{"(eq? 1 2)", `#f`},
		{"(eq? '(1 2) '(1 2))", `#f`},
		{"(equal? 'a 'a)", `#t`},
		{"(equal? 1 1)", `#t`},
		{"(equal? 1 2)", `#f`},
		{"(equal? '(1 2) '(1 2))", `#t`},
		{"(> 200 10)", `#t`},
		{"(> 10 10)", `#f`},
		{"(> 1 2)", `#f`},
		{"(>= 200 10)", `#t`},
		{"(>= 10 10)", `#t`},
		{"(>= 1 2)", `#f`},
		{"(< 200 10)", `#f`},
		{"(< 10 10)", `#f`},
		{"(< 1 2)", `#t`},
		{"(<= 200 10)", `#f`},
		{"(<= 10 10)", `#t`},
		{"(<= 1 2)", `#t`},
		{"(= 10 10)", `#t`},
		{"(= 1 2)", `#f`},
		{"(and 10 12)", `12`},
		{"(and 10 #f 20)", `#f`},
		{"(and #f undefined-proc)", `#f`},
		{"(or 10 12)", `10`},
		{"(or 10 #f 20)", `10`},
		{"(or #f #f)", `#f`},
		{"(or 1 undefined-proc)", `1`},
		{"(null? 1)", `#f`},
		{"(null? #f)", `#f`},
		{"(null? '(1))", `#f`},
		{"(null? '())", `#t`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Builtin_ConOperation(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(cons 1 2)", `(1 . 2)`},
		{"(cons 1 '())", `(1)`},
		{"(cons '(1) '(2 3 4))", `((1) 2 3 4)`},
		{"(cons \"1\" '(2 3 4))", `("1" 2 3 4)`},
		{"(cons '(1 2) 3)", `((1 2) . 3)`},
		{"(car (cons 1 2))", `1`},
		{"(cdr (cons 1 2))", `2`},
		{"(car '(1 2 3))", `1`},
		{"(cdr '(1 2 3))", `(2 3)`},
		{"(car '((1 2) (3 4)))", `(1 2)`},
		{"(cdr '((1 2) (3 4)))", `((3 4))`},
		{"(caar '((1 2)))", `1`},
		{"(cadr '(1 2 3))", `2`},
		{"(cdar '((1 2) (3 4)))", `(2)`},
		{"(caddr '((1 2) (3 4) (5 6)))", `(5 6)`},
		{"(cddr '((1 2) (3 4) (5 6)))", `((5 6))`},
		{"(cdddr '((1 2) (3 4) (5 6) (7 8)))", `((7 8))`},
		{"(caadr '(1 (2 3)))", `2`},
		{"(cdadr '(1 (2 3)))", `(3)`},
		{"(cadddr '((1 2) (3 4) (5 6) (7 8)))", `(7 8)`},
		{`(define l (list 1 2 3)) (set-car! l 4) l`, `(4 2 3)`},
		{`(define l (list 1 2 3)) (set-cdr! l 4) l`, `(1 . 4)`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Builtin_Assoc(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(assoc 1 (list '(3 2) '(2 1) '(1 9) ))", `(1 9)`},
		{"(assoc 5 (list '(3 2) '(2 1) '(1 9) ))", `#f`},
		{"(assoc 1 (cons (cons 1 2) (cons (cons 2 3) '())))", `(1 . 2)`},
		{"(assoc 5 (cons (cons 1 2) (cons (cons 2 3) '())))", `#f`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Builtin_Random(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(random 100)", `44`},
		{"(random 100.0)", `15.44158638468602`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Procedure(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(define (add a b) (+ a b)) (add 1 2)", `3`},
		{"(define (increment a) (+ a 1)) (increment 1)", `2`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_DelayAndForce(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(delay (+ 1 2))", `<promise>`},
		{"(force (delay (+ 1 2)))", `3`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Stream(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(cons-stream 1 2)", `(1 . <promise>)`},
		{"(car (cons-stream 1 2))", `1`},
		{"(cdr (cons-stream 1 2))", `<promise>`},
		{"(stream-car (cons-stream 1 2))", `1`},
		{"(stream-cdr (cons-stream 1 2))", `2`},
		{"(stream-null? (cons-stream 1 2))", `#f`},
		{"(stream-null? '())", `#t`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Apply(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(apply + '(1 2 3))", `6`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Length(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string
	}{
		{"(length '(1 2 3))", `3`},
	}

	for _, tt := range tests {
		ret := testEval(tt.input, t)
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.input, tt.expectedOutput, ret.String())
		}
	}
}

func TestEvaluator_Read(t *testing.T) {
	tests := []struct {
		stdinInput     string
		expectedOutput string
	}{
		{"(1 2 3)", `(1 2 3)`},
		{"1", `1`},
		{"foo", `'foo`},
		{"(1 2 3 (4 5 6))", `(1 2 3 (4 5 6))`},
	}

	for _, tt := range tests {
		l := lexer.New(strings.NewReader("(read)"))
		p := parser.New(l)
		program, err := p.Parse()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		evaluator := New(strings.NewReader(tt.stdinInput))
		ret, err := evaluator.Eval(program)
		if err != nil {
			t.Fatalf("stdinInput %s unexpected error: %v", tt.stdinInput, err)
		}
		if ret.String() != tt.expectedOutput {
			t.Fatalf("input %s, expected %s, got %s", tt.stdinInput, tt.expectedOutput, ret.String())
		}
	}
}

func testEval(input string, t *testing.T) *ReturnValue {
	l := lexer.New(strings.NewReader(input))
	p := parser.New(l)
	program, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	evaluator := New(strings.NewReader(""))
	result, err := evaluator.Eval(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return result
}
