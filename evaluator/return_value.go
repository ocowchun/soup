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
	default:
		return "Unknown"
	}
}

type ReturnValue struct {
	Type ValueType
	Data any
}

func (rv *ReturnValue) String() string {
	switch rv.Type {
	case NumberType:
		return fmt.Sprintf("%v", rv.Data)
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
			return fmt.Sprintf("'%s", s)
		}
		return "<invalid symbol>"
	case ListType:
		if l, ok := rv.Data.(*ListValue); ok {
			var b strings.Builder
			b.WriteString("(")
			for i, elem := range l.Elements {
				b.WriteString(elem.String())
				if i != len(l.Elements)-1 {
					b.WriteString(" ")
				}
			}
			b.WriteString(")")
			return b.String()
		}
		return "<invalid list!>"
	case ConsType:
		if c, ok := rv.Data.(*ConsValue); ok {
			var b strings.Builder
			b.WriteString("(")
			b.WriteString(c.Car.String())
			b.WriteString(" . ")
			b.WriteString(c.Cdr.String())
			b.WriteString(")")
			return b.String()
		}
		fmt.Println("invalid cons ->", rv.Data)
		return "<invalid cons>"
	default:
		return "<unknown return value type>"
	}
}

func (rv *ReturnValue) Number() float64 {
	if rv.Type != NumberType {
		panic("not a number")
	}
	if num, ok := rv.Data.(float64); ok {
		return num
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

//type SymbolValue struct {
//	Value string
//}

type ListValue struct {
	Elements []*ReturnValue
}

type ConsValue struct {
	Car *ReturnValue
	Cdr *ReturnValue
}
