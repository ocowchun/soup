package evaluator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"

	"github.com/ocowchun/soup/lexer"
)

func getCar(val *ReturnValue) (*ReturnValue, error) {
	switch val.Type {
	case ConsType:
		cons := val.Cons()
		return cons.Car, nil
	case ListType:
		list := val.List()
		if len(list.Elements) == 0 {
			return nil, fmt.Errorf("cannot call 'car' on an empty list")
		}
		return list.Elements[0], nil
	default:
		return nil, fmt.Errorf("'car' expected cons or list value, got %s", val.Type)
	}
}
func getCdr(val *ReturnValue) (*ReturnValue, error) {
	switch val.Type {
	case ConsType:
		cons := val.Cons()
		return cons.Cdr, nil
	case ListType:
		list := val.List()
		if len(list.Elements) == 0 {
			return nil, fmt.Errorf("cannot call 'cdr' on an empty list")
		}
		newList := &ListValue{Elements: list.Elements[1:]}
		return &ReturnValue{Type: ListType, Data: newList}, nil
	default:
		return nil, fmt.Errorf("'cdr' expected cons or list value, got %s", val.Type)
	}
}

type ConOperation uint8

const (
	CON_OP_CAR ConOperation = iota
	CON_OP_CDR
)

func ConProcedureFactory(operations []ConOperation) *BuiltinFunction {
	return &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("expected 1 argument, got %d", len(parameters))
			}

			val := parameters[0]
			var err error

			for _, op := range operations {
				switch op {
				case CON_OP_CAR:
					val, err = getCar(val)
					if err != nil {
						return nil, err
					}
				case CON_OP_CDR:
					val, err = getCdr(val)
					if err != nil {
						return nil, err
					}
				}
			}

			return val, nil
		},
	}
}

func isPair(val *ReturnValue) bool {
	switch val.Type {
	case ConsType:
		return true
	case ListType:
		list := val.List()
		if len(list.Elements) > 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func compareNumber(parameters []*ReturnValue, op string, evaluator *Evaluator, environment *Environment) (int, error) {
	if len(parameters) != 2 {
		return 0, fmt.Errorf("'%s' has been called with %d arguments; it requires exactly 1 argument", op, len(parameters))
	}

	left := parameters[0]
	if left.Type != NumberType {
		return 0, fmt.Errorf("!expected number value, got %s", left.Type)
	}
	leftVal := left.Number().Float64()

	right := parameters[1]
	if right.Type != NumberType {
		return 0, fmt.Errorf("expected number value, got %s", right.Type)
	}
	rightVal := right.Number().Float64()

	if leftVal > rightVal {
		return 1, nil
	} else if leftVal < rightVal {
		return -1, nil
	} else {
		return 0, nil
	}
}

func force(val *ReturnValue, evaluator *Evaluator) (*ReturnValue, error) {
	if val.Type != PromiseType {
		return nil, fmt.Errorf("expected promise type, got %s", val.Type)
	}
	promise := val.Promise()
	if promise.EvaluatedValue != nil {
		return promise.EvaluatedValue, nil
	}

	evaluatedValue, err := evaluator.eval(promise.Expression, promise.Env)
	if err != nil {
		return nil, err
	}
	promise.EvaluatedValue = evaluatedValue

	return evaluatedValue, nil

}
func isNull(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
	if len(parameters) != 1 {
		return nil, fmt.Errorf("'null?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
	}

	val := parameters[0]
	if val.Type != ListType {
		return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
	}

	if len(val.List().Elements) == 0 {
		return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
	} else {
		return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
	}
}

func initGlobalEnvironment(stdin io.Reader) *Environment {
	env := newEnvironment()
	// Add built-in functions to the environment

	//env["the-empty-stream"]
	env.Put("the-empty-stream", &ReturnValue{Type: ListType, Data: &ListValue{Elements: make([]*ReturnValue, 0)}})

	addBuiltinToEnv(env, "+", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			res := float64(0)
			for _, val := range parameters {
				if val.Type != NumberType {
					return nil, fmt.Errorf("all arguments to '+' must be numbers, got %s", val.Type)
				}
				res += val.Number().Float64()
			}
			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(res)}, nil
		},
	})

	addBuiltinToEnv(env, "-", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {

			if len(parameters) == 0 {
				return nil, fmt.Errorf("'-' requires at least one argument")
			}
			if len(parameters) == 1 {
				val := parameters[0]
				if val.Type != NumberType {
					return nil, fmt.Errorf("all arguments to '-' must be numbers, got %s", val.Type)
				}
				num := val.Number()
				if num.isInt64() && num.Int64() != math.MinInt64 {
					i := num.Int64() * -1
					return &ReturnValue{Type: NumberType, Data: MakeInt64Number(i)}, nil
				}

				return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(-val.Number().Float64())}, nil
			}

			res := float64(0)
			for i, val := range parameters {
				if val.Type != NumberType {
					return nil, fmt.Errorf("all arguments to '-' must be numbers, got %s", val.Type)
				}

				if i == 0 {
					res = val.Number().Float64()
				} else {
					res -= val.Number().Float64()
				}
			}

			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(res)}, nil
		},
	})

	addBuiltinToEnv(env, "*", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			res := float64(1)

			if len(parameters) == 0 {
				return nil, fmt.Errorf("'*' requires at least one argument")
			}

			for _, parameter := range parameters {
				if parameter.Type != NumberType {
					return nil, fmt.Errorf("all arguments to '*' must be numbers, got %s", parameter.Type)
				}
				res *= parameter.Number().Float64()
			}

			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(res)}, nil
		},
	})

	addBuiltinToEnv(env, "/", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			res := float64(0)

			if len(parameters) == 0 {
				return nil, fmt.Errorf("'/' requires at least one argument")
			}

			for i, parameter := range parameters {
				if parameter.Type != NumberType {
					return nil, fmt.Errorf("all arguments to '/' must be numbers, got %s", parameter.Type)
				}
				if i == 0 {
					res = parameter.Number().Float64()
				} else {
					res /= parameter.Number().Float64()
				}
			}

			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(res)}, nil
		},
	})

	addBuiltinToEnv(env, "remainder", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'remainder' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			a := parameters[0]
			if a.Type != NumberType {
				return nil, fmt.Errorf("expected number value, got %s", a.Type)
			}
			b := parameters[1]
			if b.Type != NumberType {
				return nil, fmt.Errorf("expected number value, got %s", b.Type)
			}

			if a.Number().isInt64() && b.Number().isInt64() {
				data := a.Number().Int64() % b.Number().Int64()
				return &ReturnValue{Type: NumberType, Data: MakeInt64Number(data)}, nil
			}
			data := math.Mod(a.Number().Float64(), b.Number().Float64())
			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(data)}, nil
		},
	})

	addBuiltinToEnv(env, "sqrt", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'sqrt' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			a := parameters[0]
			if a.Type != NumberType {
				return nil, fmt.Errorf("expected number value, got %s", a.Type)
			}
			res := math.Sqrt(a.Number().Float64())

			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(res)}, nil
		},
	})

	addBuiltinToEnv(env, "abs", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'abs' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			a := parameters[0]
			if a.Type != NumberType {
				return nil, fmt.Errorf("expected number value, got %s", a.Type)
			}

			if a.Number().isInt64() && a.Number().Int64() != math.MinInt64 {
				res := a.Number().Int64()
				if res < 0 {
					res *= -1
				}
				return &ReturnValue{Type: NumberType, Data: MakeInt64Number(res)}, nil
			}

			res := math.Abs(a.Number().Float64())

			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(res)}, nil
		},
	})

	addBuiltinToEnv(env, "number?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'number?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if val.Type == NumberType {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "string?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'string?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if val.Type == StringType {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "symbol?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'symbol?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if val.Type == SymbolType {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "pair?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'pair?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if isPair(val) {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "list?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'pair?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if val.Type == ListType {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			}
			return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
		},
	})

	// https://docs.scheme.org/schintro/schintro_49.html
	// For this, you use eq?. eq? compares two values to see if they refer to the same object.
	// Since all values in Scheme are (conceptually) pointers, this is just a pointer comparison, so eq? is always fast.
	addBuiltinToEnv(env, "eq?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'eq?' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			val1 := parameters[0]
			val2 := parameters[1]

			if val1 == val2 {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			}

			if val1.Type == val2.Type {
				switch val1.Type {
				case ConsType:
					if val1.Constant() == val2.Constant() {
						return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
					}
				case NumberType:
					if val1.Number() == val2.Number() {
						return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
					}
				case StringType:

					if val1.String() == val2.String() {
						return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
					}
				case SymbolType:
					if val1.Symbol() == val2.Symbol() {
						return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
					}
				case ListType:
					if len(val1.List().Elements) == 0 && len(val2.List().Elements) == 0 {
						return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
					}
				}
			}

			return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
		},
	})

	addBuiltinToEnv(env, "equal?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'equal?' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			val1 := parameters[0]
			val2 := parameters[1]
			if equal(val1, val2) {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, ">", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			cmp, err := compareNumber(parameters, ">", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp > 0 {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, ">=", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			cmp, err := compareNumber(parameters, ">=", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp >= 0 {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "<", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			cmp, err := compareNumber(parameters, "<", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp < 0 {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "<=", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			cmp, err := compareNumber(parameters, "<=", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp <= 0 {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "=", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			cmp, err := compareNumber(parameters, "=", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp == 0 {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			} else {
				return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
			}
		},
	})

	addBuiltinToEnv(env, "and", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			res := &ReturnValue{Type: ConstantType, Data: TrueValue}
			for _, parameter := range parameters {
				if parameter.Type == ConstantType && parameter.Constant() == FalseValue {
					return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
				}
				res = parameter
			}
			return res, nil
		},
	})

	addBuiltinToEnv(env, "or", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			for _, parameter := range parameters {
				if parameter.Type != ConstantType || parameter.Constant() != FalseValue {
					return parameter, nil
				}
			}
			return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
		},
	})

	addBuiltinToEnv(env, "not", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'cons' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if val.Type == ConstantType && val.Constant() == FalseValue {
				return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
			}
			return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
		},
	})

	// cons
	//https://groups.csail.mit.edu/mac/ftpdir/scheme-7.4/doc-html/scheme_8.html#SEC73
	addBuiltinToEnv(env, "cons", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'cons' has been called with %d arguments; it requires exactly 2 arguments", len(parameters))
			}
			car := parameters[0]
			cdr := parameters[1]
			if cdr.Type == ListType {
				cdrList := cdr.List()
				list := &ListValue{Elements: []*ReturnValue{car}}
				if len(cdrList.Elements) == 0 {
					return &ReturnValue{Type: ListType, Data: list}, nil
				}

				list.Elements = append(list.Elements, cdrList.Elements...)
				return &ReturnValue{Type: ListType, Data: list}, nil
			}

			cons := &ConsValue{
				Car: car,
				Cdr: cdr,
			}
			return &ReturnValue{Type: ConsType, Data: cons}, nil
		},
	})

	addBuiltinToEnv(env, "list", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			list := &ListValue{Elements: parameters}
			return &ReturnValue{Type: ListType, Data: list}, nil
		},
	})

	addBuiltinToEnv(env, "length", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'length' has been called with %d arguments; it requires exactly 1 arguments", len(parameters))
			}

			parameter := parameters[0]
			if parameter.Type != ListType {
				return nil, fmt.Errorf("expected list value, got %s", parameter.Type)
			}

			return &ReturnValue{Type: NumberType, Data: MakeInt64Number(int64(len(parameter.List().Elements)))}, nil
		},
	})

	addBuiltinToEnv(env, "append", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) < 2 {
				return nil, fmt.Errorf("`append` has been called with %d arguments; it requires at lesat 2 argument", len(parameters))
			}
			elements := make([]*ReturnValue, 0)
			for _, parameter := range parameters {
				if parameter.Type != ListType {
					return nil, fmt.Errorf("expected list value, got %s", parameter.Type)
				}

				elements = append(elements, parameter.List().Elements...)
			}
			list := &ListValue{Elements: elements}
			return &ReturnValue{Type: ListType, Data: list}, nil
		},
	})

	addBuiltinToEnv(env, "car", ConProcedureFactory([]ConOperation{CON_OP_CAR}))
	addBuiltinToEnv(env, "cdr", ConProcedureFactory([]ConOperation{CON_OP_CDR}))
	addBuiltinToEnv(env, "caar", ConProcedureFactory([]ConOperation{CON_OP_CAR, CON_OP_CAR}))
	addBuiltinToEnv(env, "cadr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CAR}))
	addBuiltinToEnv(env, "cddr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CDR}))
	addBuiltinToEnv(env, "cdar", ConProcedureFactory([]ConOperation{CON_OP_CAR, CON_OP_CDR}))
	addBuiltinToEnv(env, "caddr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CDR, CON_OP_CAR}))
	addBuiltinToEnv(env, "caadr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CAR, CON_OP_CAR}))
	addBuiltinToEnv(env, "cdadr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CAR, CON_OP_CDR}))
	addBuiltinToEnv(env, "cdddr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CDR, CON_OP_CDR}))
	addBuiltinToEnv(env, "cadddr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CDR, CON_OP_CDR, CON_OP_CAR}))

	//https://groups.csail.mit.edu/mac/ftpdir/scheme-7.4/doc-html/scheme_8.html
	addBuiltinToEnv(env, "set-car!", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'set-car!' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			carVal := parameters[1]

			container := parameters[0]
			switch container.Type {
			case ConsType:
				cons := container.Cons()
				cons.Car = carVal
			case ListType:
				list := container.List()
				if len(list.Elements) == 0 {
					return nil, errors.New("cannot set-car! on an empty list")
				}
				list.Elements[0] = carVal
			default:
				return nil, fmt.Errorf("first argument to 'set-car!' must be a cons cell or a non-empty list, got %T", container)
			}

			return &ReturnValue{Type: ConstantType, Data: VoidConst}, nil
		},
	})

	addBuiltinToEnv(env, "set-cdr!", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'set-cdr!' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			cdrVal := parameters[1]

			container := parameters[0]
			switch container.Type {
			case ConsType:
				cons := container.Cons()
				cons.Cdr = cdrVal
			case ListType:
				list := container.List()
				if len(list.Elements) == 0 {
					return nil, errors.New("cannot set-cdr! on an empty list")
				}
				cons := &ConsValue{
					Car: list.Elements[0],
					Cdr: cdrVal,
				}
				container.Type = ConsType
				container.Data = cons
			default:
				return nil, fmt.Errorf("first argument to 'set-cdr!' must be a cons cell or a non-empty list, got %T", container)
			}

			return &ReturnValue{Type: ConstantType, Data: VoidConst}, nil
		},
	})

	addBuiltinToEnv(env, "stream-car", ConProcedureFactory([]ConOperation{CON_OP_CAR}))
	addBuiltinToEnv(env, "stream-cdr", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'stream-cdr' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if val.Type != ConsType {
				return nil, fmt.Errorf("first argument to 'stream-cdr' must be a cons , got %T", val.Type)
			}

			return force(val.Cons().Cdr, evaluator)
		},
	})

	addBuiltinToEnv(env, "stream-null?", &BuiltinFunction{
		Fn: isNull,
	})

	addBuiltinToEnv(env, "null?", &BuiltinFunction{
		Fn: isNull,
	})

	addBuiltinToEnv(env, "display", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'display' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]

			if val.Type == StringType {
				fmt.Print(val.StringValue())
			} else {
				fmt.Print(val.String())
			}

			return &ReturnValue{Type: ConstantType, Data: VoidConst}, nil
		},
	})

	addBuiltinToEnv(env, "newline", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 0 {
				return nil, fmt.Errorf("'newline' has been called with %d arguments; it requires exactly 0 argument", len(parameters))
			}

			fmt.Println()

			return &ReturnValue{Type: ConstantType, Data: VoidConst}, nil
		},
	})

	addBuiltinToEnv(env, "print", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'print' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			fmt.Println(val.String())

			return &ReturnValue{Type: ConstantType, Data: VoidConst}, nil
		},
	})

	// https://docs.scheme.org/schintro/schintro_69.html
	addBuiltinToEnv(env, "apply", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) < 2 {
				return nil, fmt.Errorf("'apply' has been called with %d arguments; it requires at least 2 arguments", len(parameters))
			}
			// TODO: actually I don't know the point of 3rd and later arguments, current implementation simply skip those arguments

			proc := parameters[0]
			list := parameters[1]
			if list.Type != ListType {
				return nil, fmt.Errorf("'apply' expect second argument to be list but got %s", list.Type)
			}

			switch proc.Type {
			case BuiltinFunctionType:
				fn := proc.BuiltinFunction()
				return evaluator.evalBuiltinFunction(fn, list.List().Elements, environment)
			case ProcedureType:
				fn := proc.Procedure()
				return evaluator.evalProcedure(fn, list.List().Elements, environment)
			default:
				return nil, fmt.Errorf("'apply' expect first argument to be procedure/builtinFunction but got %s", list.Type)
			}
		},
	})

	// TODO implement assoc, map
	addBuiltinToEnv(env, "map", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) < 2 {
				return nil, fmt.Errorf("'assoc' has been called with %d arguments; it requires at least 2 arguments", len(parameters))
			}

			proc := parameters[0]

			operandsList := make([][]*ReturnValue, 0)
			for i := 1; i < len(parameters); i++ {
				val := parameters[i]
				if val.Type == ListType {
					list := val.List()
					if i == 1 {
						for _, element := range list.Elements {
							operandsList = append(operandsList, []*ReturnValue{element})
						}
						continue
					}
					if len(list.Elements) != len(operandsList) {
						return nil, fmt.Errorf("all lists must have same size")
					}

					for j, element := range list.Elements {
						operandsList[j] = append(operandsList[j], element)
					}
				} else {
					return nil, fmt.Errorf("expect parameter to be list but got %s", val.Type)
				}
			}

			res := make([]*ReturnValue, 0)
			switch proc.Type {
			case BuiltinFunctionType:
				procedure := proc.BuiltinFunction()
				for _, operands := range operandsList {
					ret, err := evaluator.evalBuiltinFunction(procedure, operands, environment)
					if err != nil {
						return nil, err
					}
					res = append(res, ret)
				}
			case ProcedureType:
				procedure := proc.Procedure()
				for _, operands := range operandsList {
					ret, err := evaluator.evalProcedure(procedure, operands, environment)
					if err != nil {
						return nil, err
					}
					res = append(res, ret)
				}
			default:
				return nil, fmt.Errorf("unknown procedure type %s", proc.Type)
			}

			return &ReturnValue{Type: ListType, Data: &ListValue{Elements: res}}, nil
		},
	})
	addBuiltinToEnv(env, "assoc", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'assoc' has been called with %d arguments; it requires exactly 2 arguments", len(parameters))
			}
			key := parameters[0]
			val := parameters[1]

			if val.Type == ListType {
				list := val.List()
				for _, item := range list.Elements {
					switch item.Type {
					case ConsType:
						pair := item.Cons()
						if equal(pair.Car, key) {
							return item, nil
						}
					case ListType:
						pairList := item.List()
						if len(pairList.Elements) == 0 {
							return nil, fmt.Errorf("non-pair found in list")
						}
						if equal(pairList.Elements[0], key) {
							return item, nil
						}
					default:
						return nil, fmt.Errorf("non-pair found in list")
					}
				}
			} else if val.Type == ConsType {
				currentCons := val.Cons()
				for {
					switch currentCons.Car.Type {
					case ConsType:
						cons := currentCons.Car.Cons()
						if equal(cons.Car, key) {
							return currentCons.Car, nil
						}
					case ListType:
						pairList := currentCons.Car.List()
						if len(pairList.Elements) == 0 {
							return nil, fmt.Errorf("non-pair found in list")
						}
						if equal(pairList.Elements[0], key) {
							return currentCons.Car, nil
						}
					default:
						return nil, fmt.Errorf("non-pair found in list, type is %s", currentCons.Car.Type)
					}
					if currentCons.Cdr.Type == ConsType {
						currentCons = currentCons.Cdr.Cons()
					} else {
						break
					}
				}
			} else {
				return nil, fmt.Errorf("expected list value, got %s", val.Type)
			}

			return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
		},
	})

	addBuiltinToEnv(env, "error", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) < 1 {
				return nil, fmt.Errorf("'error' has been called with %d arguments; it requires at least 1 argument", len(parameters))
			}

			val := parameters[0]
			if len(parameters) == 1 {
				return nil, fmt.Errorf("failed to evaluate: %s", val.String())

			}
			val2 := parameters[1]
			return nil, fmt.Errorf("failed to evaluate: %s, %s", val.String(), val2.String())
		},
	})

	r := rand.New(rand.NewSource(9527))
	// https://groups.csail.mit.edu/mac/ftpdir/scheme-7.4/doc-html/scheme_5.html#SEC53
	addBuiltinToEnv(env, "random", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			// TODO: implement random-state
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'random' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val := parameters[0]
			if val.Type != NumberType {
				return nil, fmt.Errorf("expected number type, got %s", val.Type)
			}

			if val.Number().isInt64() {
				res := r.Int63n(val.Number().Int64())
				return &ReturnValue{Type: NumberType, Data: MakeInt64Number(res)}, nil
			}

			res := r.Float64() * val.Number().Float64()
			return &ReturnValue{Type: NumberType, Data: MakeFloat64Number(res)}, nil
		},
	})

	addBuiltinToEnv(env, "force", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'force' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			return force(parameters[0], evaluator)
		},
	})

	//https: //docs.scheme.org/schintro/schintro_115.html#SEC135
	addBuiltinToEnv(env, "read", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 0 {
				return nil, fmt.Errorf("'read' has been called with %d arguments; it requires exactly 0 argument", len(parameters))
			}

			reader := bufio.NewReader(stdin)
			return read(reader)
		},
	})

	// Add more built-in functions as needed
	return env
}

func readList(l *lexer.Lexer) (*ListValue, error) {
	list := &ListValue{Elements: make([]*ReturnValue, 0)}
	for {
		tok := l.NextToken()
		if tok.TokenType == lexer.TokenTypeRightParen {
			return list, nil
		} else if tok.TokenType == lexer.TokenTypeLeftParen {
			subList, err := readList(l)
			if err != nil {
				return nil, err
			}
			list.Elements = append(list.Elements, &ReturnValue{Type: ListType, Data: subList})
		} else if tok.TokenType == lexer.TokenTypeNumber {
			num, err := MakeNumber(tok.Content)
			if err != nil {
				return nil, err
			}
			list.Elements = append(list.Elements, num)
		} else if tok.TokenType == lexer.TokenTypeEOF || tok.TokenType == lexer.TokenTypeInvalid {
			panic("unreachable")
		} else if tok.TokenType == lexer.TokenTypeQuote {
			// how to handle this case?
			head := &ReturnValue{Type: SymbolType, Data: "quote"}
			inner, err := doRead(l)
			if err != nil {
				return nil, err
			}

			element := &ReturnValue{Type: ListType, Data: &ListValue{Elements: []*ReturnValue{head, inner}}}
			list.Elements = append(list.Elements, element)
		} else {
			sym := &ReturnValue{Type: SymbolType, Data: tok.Content}
			list.Elements = append(list.Elements, sym)
		}
	}
}

func read(reader io.Reader) (*ReturnValue, error) {
	l := lexer.New(reader)
	return doRead(l)
}

func doRead(l *lexer.Lexer) (*ReturnValue, error) {
	firstToken := l.NextToken()
	if firstToken.TokenType == lexer.TokenTypeRightParen {
		return nil, fmt.Errorf("unexpected ')'")
	} else if firstToken.TokenType == lexer.TokenTypeLeftParen {
		list, err := readList(l)
		if err != nil {
			return nil, err
		}
		return &ReturnValue{Type: ListType, Data: list}, nil
	} else if firstToken.TokenType == lexer.TokenTypeNumber {
		return MakeNumber(firstToken.Content)
	} else if firstToken.TokenType == lexer.TokenTypeQuote {
		head := &ReturnValue{Type: SymbolType, Data: "quote"}
		tail, err := doRead(l)
		if err != nil {
			return nil, err
		}
		list := &ListValue{Elements: []*ReturnValue{head, tail}}
		return &ReturnValue{Type: ListType, Data: list}, nil
	} else {
		return &ReturnValue{Type: SymbolType, Data: firstToken.Content}, nil
	}
}
