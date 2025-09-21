package evaluator

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
)

func car(val *ReturnValue) (*ReturnValue, error) {
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
		return nil, fmt.Errorf("expected cons or list value, got %s", val.Type)
	}
}
func cdr(val *ReturnValue) (*ReturnValue, error) {
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
		return nil, fmt.Errorf("expected cons or list value, got %s", val.Type)
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
					val, err = car(val)
					if err != nil {
						return nil, err
					}
				case CON_OP_CDR:
					val, err = cdr(val)
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

func initGlobalEnvironment() *Environment {
	env := newEnvironment()
	// Add built-in functions to the environment

	addBuiltinToEnv(env, "+", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			res := float64(0)
			for _, val := range parameters {
				if val.Type != NumberType {
					return nil, errors.New("all arguments to '+' must be numbers")
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
					return nil, errors.New("all arguments to '-' must be numbers")
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
					return nil, errors.New("all arguments to '-' must be numbers")
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
					return nil, errors.New("all arguments to '*' must be numbers")
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

			if val1.Type == NumberType && val2.Type == NumberType {
				if val1.Number() == val2.Number() {
					return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
				}
			}
			if val1.Type == StringType && val2.Type == StringType {
				if val1.String() == val2.String() {
					return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
				}
			}
			if val1.Type == SymbolType && val2.Type == SymbolType {
				if val1.Symbol() == val2.Symbol() {
					return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
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
	addBuiltinToEnv(env, "cons", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'cons' has been called with %d arguments; it requires exactly 2 arguments", len(parameters))
			}

			cons := &ConsValue{
				Car: parameters[0],
				Cdr: parameters[1],
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
	addBuiltinToEnv(env, "cdar", ConProcedureFactory([]ConOperation{CON_OP_CAR, CON_OP_CDR}))
	addBuiltinToEnv(env, "caddr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CDR, CON_OP_CAR}))
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

	addBuiltinToEnv(env, "null?", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
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
		},
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
				//panic(fmt.Sprintf("failed to evaluate: %s", val.String()))
				return nil, fmt.Errorf("failed to evaluate: %s", val.String())

			}
			val2 := parameters[1]
			return nil, fmt.Errorf("failed to evaluate: %s, %s", val.String(), val2.String())
		},
	})

	r := rand.New(rand.NewSource(9527))
	// TODO: implement random
	// https://groups.csail.mit.edu/mac/ftpdir/scheme-7.4/doc-html/scheme_5.html#SEC53
	addBuiltinToEnv(env, "random", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			// TODO: implement random-state
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'random' has been called with %d arguments; it exactly 1 argument", len(parameters))
			}

			// how to handle int64 and float64

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

	// Add more built-in functions as needed
	return env
}
