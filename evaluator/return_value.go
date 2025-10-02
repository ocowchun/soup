package evaluator

import (
	"fmt"
	"strings"

	"github.com/ocowchun/soup/parser"
)

type ValueType uint8

const (
	NumberType ValueType = iota
	StringType
	ConstantType
	ProcedureType
	BuiltinFunctionType
	SymbolType
	ListType
	ConsType
	PromiseType
)

func (t ValueType) String() string {
	switch t {
	case NumberType:
		return "Number"
	case StringType:
		return "String"
	case ConstantType:
		return "Constant"
	case ProcedureType:
		return "Procedure"
	case BuiltinFunctionType:
		return "BuiltinFunction"
	case SymbolType:
		return "Symbol"
	case ListType:
		return "List"
	case ConsType:
		return "Cons"
	case PromiseType:
		return "Promise"
	default:
		return "Unknown"
	}
}

type ReturnValue struct {
	Type ValueType
	Data any
}

func (rv *ReturnValue) String() string {
	return rv.Display(0)
}

func (rv *ReturnValue) Display(depth int) string {
	switch rv.Type {
	case NumberType:
		if c, ok := rv.Data.(Number); ok {
			return c.String()
		} else {
			return "<invalid number>"
		}
	case StringType:
		return fmt.Sprintf("\"%s\"", rv.Data)
	case ConstantType:
		if c, ok := rv.Data.(ConstantValue); ok {
			return c.String()
		} else {
			return "<invalid constant>"
		}
	case ProcedureType:
		return "<procedure>"
	case BuiltinFunctionType:
		return "<builtin function>"
	case SymbolType:
		if s, ok := rv.Data.(string); ok {
			if depth == 0 {
				return fmt.Sprintf("'%s", s)
			} else {
				return fmt.Sprintf("%s", s)
			}
		}
		return "<invalid symbol>"
	case ListType:
		l, ok := rv.Data.(*ListValue)
		if !ok {
			return "<invalid list!>"
		}

		var b strings.Builder
		if depth == 0 {
			b.WriteString("'")
		}
		elements := l.Elements
		if len(elements) == 2 && elements[0].Type == SymbolType && elements[0].Symbol() == "quote" {
			if elements[1].Type == SymbolType && elements[1].Symbol() != "quote" {
				return fmt.Sprintf("''%s", elements[1].Symbol())
			} else if depth > 0 && elements[1].Type == ListType {
				return fmt.Sprintf("'%s", elements[1].Display(depth+1))
			} else {
				return fmt.Sprintf("''%s", elements[1].Display(depth+1))
			}
		}

		b.WriteString("(")
		for i, elem := range elements {
			b.WriteString(elem.Display(depth + 1))
			if i != len(elements)-1 {
				b.WriteString(" ")
			}
		}
		b.WriteString(")")
		return b.String()
	case ConsType:
		c, ok := rv.Data.(*ConsValue)
		if !ok {
			return "<invalid cons>"
		}

		var b strings.Builder
		if depth == 0 {
			b.WriteString("'")
		}

		b.WriteString("(")
		b.WriteString(c.Car.Display(depth + 1))
		b.WriteString(" . ")
		b.WriteString(c.Cdr.Display(depth + 1))
		b.WriteString(")")
		return b.String()
	case PromiseType:
		return "<promise>"
	default:
		return "<unknown return value type>"
	}
}

func (rv *ReturnValue) Number() Number {
	if rv.Type != NumberType {
		panic("not a number")
	}
	if n, ok := rv.Data.(Number); ok {
		return n
	}
	panic("invalid number")
}

func (rv *ReturnValue) StringValue() string {
	if rv.Type != StringType {
		panic("not a string")
	}
	if str, ok := rv.Data.(string); ok {
		return str
	}
	panic("invalid string")
}

func (rv *ReturnValue) Constant() ConstantValue {
	if rv.Type != ConstantType {
		panic("not a constant")
	}
	if c, ok := rv.Data.(ConstantValue); ok {
		return c
	}
	panic("invalid constant")
}

func (rv *ReturnValue) Procedure() *ProcedureValue {
	if rv.Type != ProcedureType {
		panic("not a procedure")
	}
	if proc, ok := rv.Data.(*ProcedureValue); ok {
		return proc
	}
	panic("invalid procedure")
}

func (rv *ReturnValue) BuiltinFunction() *BuiltinFunction {
	if rv.Type != BuiltinFunctionType {
		panic("not a builtin function")
	}
	if fn, ok := rv.Data.(*BuiltinFunction); ok {
		return fn
	}
	panic("invalid builtin function")
}

func (rv *ReturnValue) Symbol() string {
	if rv.Type != SymbolType {
		panic("not a symbol")
	}
	if str, ok := rv.Data.(string); ok {
		return str
	}
	panic("invalid symbol")
}

func (rv *ReturnValue) List() *ListValue {
	if rv.Type != ListType {
		panic("not a list")
	}
	if list, ok := rv.Data.(*ListValue); ok {
		return list
	}
	panic("invalid list")
}

func (rv *ReturnValue) Cons() *ConsValue {
	if rv.Type != ConsType {
		panic("not a cons")
	}
	if cons, ok := rv.Data.(*ConsValue); ok {
		return cons
	}
	panic("invalid cons")
}

func (rv *ReturnValue) Promise() *PromiseValue {
	if rv.Type != PromiseType {
		panic("not a promise")
	}
	if promise, ok := rv.Data.(*PromiseValue); ok {
		return promise
	}
	panic("invalid promise")
}

type ConstantValue uint8

const (
	VoidConst ConstantValue = iota
	TrueValue
	FalseValue
)

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

func (p *ProcedureValue) CaneTakeArbitraryParameters() bool {
	return p.OptionalTailParameter != ""
}

type BuiltinFunction struct {
	//Fn func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (*ReturnValue, error)
	Fn func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error)
}

type ListValue struct {
	Elements []*ReturnValue
}

type ConsValue struct {
	Car *ReturnValue
	Cdr *ReturnValue
}

type PromiseValue struct {
	Expression     parser.Expression
	Env            *Environment
	EvaluatedValue *ReturnValue
}
