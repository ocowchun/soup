package evaluator

import (
	"errors"
	"fmt"
	"math"
)

func car(val *ReturnValue) (*ReturnValue, error) {
	switch val.Type {
	case ConsType:
		cons := val.Cons()
		return cons.Car, nil
	case ListType:
		list := val.List()
		if len(list.Elements) == 0 {
			return nil, fmt.Errorf("cannot call 'cdr' on an empty list")
		}
		return list.Elements[0], nil
	default:
		return nil, fmt.Errorf("expected cons or list value, got %T", val)
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
		return nil, fmt.Errorf("expected cons or list value, got %T", val)
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
				res += val.Number()
			}
			return &ReturnValue{Type: NumberType, Data: res}, nil
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
				//return &NumberValue{Value: -numVal.Value}, nil
				return &ReturnValue{Type: NumberType, Data: -val.Number()}, nil
			}

			res := float64(0)
			for i, val := range parameters {
				if val.Type != NumberType {
					return nil, errors.New("all arguments to '-' must be numbers")
				}

				if i == 0 {
					res = val.Number()
				} else {
					res -= val.Number()
				}
			}

			return &ReturnValue{Type: NumberType, Data: res}, nil
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
				res *= parameter.Number()
			}

			return &ReturnValue{Type: NumberType, Data: res}, nil
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
					res = parameter.Number()
				} else {
					res /= parameter.Number()
				}
			}

			return &ReturnValue{Type: NumberType, Data: res}, nil
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
			res := math.Mod(a.Number(), b.Number())

			return &ReturnValue{Type: NumberType, Data: res}, nil
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
			res := math.Sqrt(a.Number())

			return &ReturnValue{Type: NumberType, Data: res}, nil
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
			res := math.Abs(a.Number())

			return &ReturnValue{Type: NumberType, Data: res}, nil
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

	// cons
	addBuiltinToEnv(env, "cons", &BuiltinFunction{
		Fn: func(parameters []*ReturnValue, evaluator *Evaluator, environment *Environment) (*ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'cons' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
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
				panic(fmt.Sprintf("failed to evaluate: %s", val.String()))

			}
			val2 := parameters[1]
			panic(fmt.Sprintf("failed to evaluate: %s, %s", val.String(), val2.String()))
		},
	})

	// Add more built-in functions as needed
	return env
}
