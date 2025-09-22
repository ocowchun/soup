package parser

import (
	"fmt"
	"strings"

	"github.com/ocowchun/soup/lexer"
)

type Expression interface {
	expressionNode()
	String() string
	Token() lexer.Token
}

type NumberLiteral struct {
	NumToken lexer.Token
}

func (n *NumberLiteral) expressionNode() {}
func (n *NumberLiteral) String() string {
	return fmt.Sprintf("%v", n.NumToken.Content)
}
func (n *NumberLiteral) Token() lexer.Token {
	return n.NumToken
}

type StringLiteral struct {
	StrToken lexer.Token
	Value    string
}

func (s *StringLiteral) expressionNode() {
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}
func (s *StringLiteral) Token() lexer.Token {
	return s.StrToken
}

type CallExpression struct {
	LeftParenToken lexer.Token
	Operator       Expression
	Operands       []Expression
}

func (a *CallExpression) expressionNode() {}
func (a *CallExpression) String() string {
	var b strings.Builder
	b.WriteString("(")
	b.WriteString(a.Operator.String())
	b.WriteString(" ")
	for i, op := range a.Operands {
		b.WriteString(op.String())
		if i != len(a.Operands)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString(")")

	return b.String()
}
func (a *CallExpression) Token() lexer.Token {
	return a.LeftParenToken
}

type PrimitiveProcedureExpression struct {
	NameToken lexer.Token
	Value     string
}

func (p *PrimitiveProcedureExpression) expressionNode() {}
func (p *PrimitiveProcedureExpression) String() string {
	return p.Value
}
func (p *PrimitiveProcedureExpression) Token() lexer.Token {
	return p.NameToken
}

type IdentifierExpression struct {
	NameToken lexer.Token
	Value     string
}

func (i *IdentifierExpression) expressionNode() {}
func (i *IdentifierExpression) String() string {
	return i.Value
}
func (i *IdentifierExpression) Token() lexer.Token {
	return i.NameToken
}

type IfExpression struct {
	LeftParenToken lexer.Token
	Predicate      Expression
	Consequent     Expression
	Alternative    Expression
}

func (i *IfExpression) expressionNode() {}
func (i *IfExpression) String() string {
	return fmt.Sprintf("(if %s %s %s)", i.Predicate.String(), i.Consequent.String(), i.Alternative.String())
}
func (i *IfExpression) Token() lexer.Token {
	return i.LeftParenToken
}

type LambdaExpression struct {
	LeftParenToken        lexer.Token
	Parameters            []string
	OptionalTailParameter string // empty if not present
	Body                  []Expression
}

func (l *LambdaExpression) expressionNode() {}
func (l *LambdaExpression) String() string {
	var b strings.Builder
	b.WriteString("(lambda (")
	for i, param := range l.Parameters {
		b.WriteString(param)
		if i != len(l.Parameters)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString(")")
	for _, expr := range l.Body {
		b.WriteString(" ")
		b.WriteString(expr.String())
	}
	b.WriteString(")")
	return b.String()
}
func (l *LambdaExpression) Token() lexer.Token {
	return l.LeftParenToken
}

type DefineExpression struct {
	LeftParenToken lexer.Token
	Name           string
	Value          Expression
}

func (d *DefineExpression) expressionNode() {}

func (d *DefineExpression) String() string {
	if lambda, ok := d.Value.(*LambdaExpression); ok {
		var b strings.Builder
		b.WriteString("(define ")
		b.WriteString("(")
		b.WriteString(d.Name)
		if len(lambda.Parameters) > 0 {
			for _, param := range lambda.Parameters {
				b.WriteString(" ")
				b.WriteString(param)
			}
		}
		b.WriteString(")")

		for _, expr := range lambda.Body {
			b.WriteString(" ")
			b.WriteString(expr.String())
		}
		b.WriteString(")")

		return b.String()
	} else {
		return fmt.Sprintf("(define %s %s)", d.Name, d.Value.String())
	}
}
func (d *DefineExpression) Token() lexer.Token {
	return d.LeftParenToken
}

type ListExpression struct {
	LeftParenToken lexer.Token
	Elements       []Expression
}

func (l *ListExpression) expressionNode() {}

func (l *ListExpression) String() string {
	var b strings.Builder
	b.WriteString("'(")
	for i, elem := range l.Elements {
		b.WriteString(elem.String())
		if i != len(l.Elements)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString(")")
	return b.String()
}
func (l *ListExpression) Token() lexer.Token {
	return l.LeftParenToken
}

type SymbolExpression struct {
	ValueToken lexer.Token
	Value      string
}

func (s *SymbolExpression) expressionNode() {}
func (s *SymbolExpression) String() string {
	return fmt.Sprintf("'%s", s.Value)
}
func (s *SymbolExpression) Token() lexer.Token {
	return s.ValueToken
}

type BeginExpression struct {
	LeftParenToken lexer.Token
	Expressions    []Expression
}

func (b *BeginExpression) expressionNode() {}
func (b *BeginExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(begin")
	for _, expr := range b.Expressions {
		sb.WriteString(" ")
		sb.WriteString(expr.String())
	}
	sb.WriteString(")")
	return sb.String()
}
func (b *BeginExpression) Token() lexer.Token {
	return b.LeftParenToken
}

type SetExpression struct {
	LeftParenToken lexer.Token
	Name           string
	Value          Expression
}

func (s *SetExpression) expressionNode() {}
func (s *SetExpression) String() string {
	return fmt.Sprintf("(set! %s %s)", s.Name, s.Value.String())
}
func (s *SetExpression) Token() lexer.Token {
	return s.LeftParenToken
}

type voidExpression struct{}

func (v *voidExpression) expressionNode() {}
func (v *voidExpression) String() string {
	return ""
}
func (v *voidExpression) Token() lexer.Token {
	panic("fix it later")
}

var Void = &voidExpression{}

type booleanLiteral struct {
	Value bool
}

func (b *booleanLiteral) expressionNode() {}
func (b *booleanLiteral) String() string {
	if b.Value {
		return "#t"
	} else {
		return "#f"
	}
}
func (b *booleanLiteral) Token() lexer.Token {
	panic("fix it later")
}

var TrueLiteral = &booleanLiteral{Value: true}
var FalseLiteral = &booleanLiteral{Value: false}

type DelayExpression struct {
	DelayToken lexer.Token
	Expression Expression
}

func (d *DelayExpression) expressionNode() {}
func (d *DelayExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(delay ")
	sb.WriteString(d.Expression.String())
	sb.WriteString(")")
	return sb.String()
}

func (d *DelayExpression) Token() lexer.Token {
	return d.DelayToken
}

type StreamExpression struct {
	ConsStreamToken lexer.Token
	CarExpression   Expression
	CdrExpression   Expression
}

func (s *StreamExpression) expressionNode() {}
func (s *StreamExpression) String() string {
	var sb strings.Builder
	sb.WriteString("(cons-stream ")
	sb.WriteString(s.CarExpression.String())
	sb.WriteString(" ")
	sb.WriteString(s.CdrExpression.String())
	sb.WriteString(")")
	return sb.String()
}

func (s *StreamExpression) Token() lexer.Token {
	return s.ConsStreamToken
}
