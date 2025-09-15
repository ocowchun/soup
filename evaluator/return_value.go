package evaluator

import (
	"fmt"
	"strings"

	"github.com/ocowchun/soup/parser"
)

type ReturnValue interface {
	returnValue()
	String() string
}

type NumberValue struct {
	Value float64
}

func (n *NumberValue) returnValue() {}
func (n *NumberValue) String() string {
	return fmt.Sprintf("%v", n.Value)
}

type StringValue struct {
	Value string
}

func (s *StringValue) returnValue() {}
func (s *StringValue) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

type ConstantValue uint8

const (
	VoidConst ConstantValue = iota
	TrueValue
	FalseValue
)

func (c ConstantValue) returnValue() {}
func (c ConstantValue) String() string {
	switch c {
	case VoidConst:
		return "<void>"
	case TrueValue:
		return "#t"
	case FalseValue:
		return "#f"
	default:
		return "<unknown constant>"
	}
}

type ProcedureValue struct {
	Parameters            []string
	OptionalTailParameter string // empty if not present
	Body                  []parser.Expression
	Env                   *Environment
}

func (p *ProcedureValue) returnValue() {}
func (p *ProcedureValue) String() string {
	return "<procedure>"
}
func (p *ProcedureValue) CaneTakeArbitraryParameters() bool {
	return p.OptionalTailParameter != ""
}

type BuiltinFunction struct {
	Fn func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error)
}

func (b *BuiltinFunction) returnValue() {}
func (b *BuiltinFunction) String() string {
	return "<builtin function>"
}

type SymbolValue struct {
	Value string
}

func (s *SymbolValue) returnValue() {}
func (s *SymbolValue) String() string {
	return fmt.Sprintf("'%s", s.Value)
}

type ListValue struct {
	Elements []ReturnValue
}

func (l *ListValue) returnValue() {}

func (l *ListValue) String() string {
	var b strings.Builder
	b.WriteString("(")
	for i, elem := range l.Elements {
		if symbol, ok := elem.(*SymbolValue); ok {
			b.WriteString(symbol.Value)
		} else {
			b.WriteString(elem.String())
		}
		if i != len(l.Elements)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString(")")
	return b.String()
}

type ConsValue struct {
	Car ReturnValue
	Cdr ReturnValue
}

func (l *ConsValue) returnValue() {}
func (l *ConsValue) String() string {
	var b strings.Builder
	b.WriteString("(")
	b.WriteString(l.Car.String())
	b.WriteString(" . ")
	b.WriteString(l.Cdr.String())
	b.WriteString(")")
	return b.String()
}
