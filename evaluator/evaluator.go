package evaluator

import (
	"errors"
	"fmt"
	"math"

	"github.com/ocowchun/soup/lexer"
	"github.com/ocowchun/soup/parser"
)

type Evaluator struct {
	globalEnv *Environment
}

func New() *Evaluator {
	env := initGlobalEnvironment()
	return &Evaluator{globalEnv: env}
}

func initGlobalEnvironment() *Environment {
	env := newEnvironment()
	// Add built-in functions to the environment
	env.Put("+", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			res := float64(0)
			for _, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				numVal, ok := val.(*NumberValue)
				if !ok {
					return nil, fmt.Errorf("expected number value, got %T", val)
				}
				res += numVal.Value
			}

			return &NumberValue{Value: res}, nil
		},
	})

	env.Put("-", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {

			if len(parameters) == 0 {
				return nil, fmt.Errorf("'-' requires at least one argument")
			}
			if len(parameters) == 1 {
				val, err := evaluator.eval(parameters[0], environment)
				if err != nil {
					return nil, err
				}
				numVal, ok := val.(*NumberValue)
				if !ok {
					return nil, fmt.Errorf("expected number value, got %T", val)
				}
				return &NumberValue{Value: -numVal.Value}, nil
			}

			res := float64(0)
			for i, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				numVal, ok := val.(*NumberValue)
				if !ok {
					return nil, fmt.Errorf("expected number value, got %T", val)
				}

				if i == 0 {
					res = numVal.Value
				} else {
					res -= numVal.Value
				}
			}

			return &NumberValue{Value: res}, nil
		},
	})

	env.Put("*", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			res := float64(1)

			if len(parameters) == 0 {
				return nil, fmt.Errorf("'*' requires at least one argument")
			}

			for _, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				numVal, ok := val.(*NumberValue)
				if !ok {
					return nil, fmt.Errorf("expected number value, got %T %s", val, parameter.String())
				}
				res *= numVal.Value
			}

			return &NumberValue{Value: res}, nil
		},
	})

	env.Put("/", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			res := float64(0)

			if len(parameters) == 0 {
				return nil, fmt.Errorf("'/' requires at least one argument")
			}

			for i, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				numVal, ok := val.(*NumberValue)
				if !ok {
					return nil, fmt.Errorf("expected number value, got %T", val)
				}
				if i == 0 {
					res = numVal.Value
				} else {
					res /= numVal.Value
				}
			}

			return &NumberValue{Value: res}, nil
		},
	})

	env.Put("remainder", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'remainder' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			a, err := evaluator.evalNumber(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			b, err := evaluator.evalNumber(parameters[1], environment)
			if err != nil {
				return nil, err
			}
			res := math.Mod(a.Value, b.Value)

			return &NumberValue{Value: res}, nil
		},
	})

	env.Put("sqrt", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'sqrt' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			numVal, ok := val.(*NumberValue)
			if !ok {
				return nil, fmt.Errorf("expected number value, got %T", val)
			}
			res := math.Sqrt(numVal.Value)

			return &NumberValue{Value: res}, nil
		},
	})

	env.Put("abs", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'abs' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			numVal, ok := val.(*NumberValue)
			if !ok {
				return nil, fmt.Errorf("expected number value, got %T", val)
			}
			res := math.Abs(numVal.Value)

			return &NumberValue{Value: res}, nil
		},
	})

	env.Put("number?", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'number?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			_, ok := val.(*NumberValue)
			if ok {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("symbol?", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'symbol?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			_, ok := val.(*SymbolValue)
			if ok {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("pair?", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'pair?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			switch val.(type) {
			case *ConsValue:
				return TrueValue, nil
			case *ListValue:
				list := val.(*ListValue)
				if len(list.Elements) > 0 {
					return TrueValue, nil
				}
				return FalseValue, nil
			default:
				return FalseValue, nil
			}
		},
	})

	// https://docs.scheme.org/schintro/schintro_49.html
	// For this, you use eq?. eq? compares two values to see if they refer to the same object.
	// Since all values in Scheme are (conceptually) pointers, this is just a pointer comparison, so eq? is always fast.
	env.Put("eq?", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'eq?' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			val1, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			val2, err := evaluator.eval(parameters[1], environment)
			if err != nil {
				return nil, err
			}

			if val1 == val2 {
				return TrueValue, nil
			}

			if num1, ok := val1.(*NumberValue); ok {
				if num2, ok := val2.(*NumberValue); ok {
					if num1.Value == num2.Value {
						return TrueValue, nil
					}
				}
			}
			if str1, ok := val1.(*SymbolValue); ok {
				if str2, ok := val2.(*SymbolValue); ok {
					if str1.Value == str2.Value {
						return TrueValue, nil
					}
				}
			}
			return FalseValue, nil
		},
	})

	env.Put(">", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			cmp, err := compareNumber(parameters, ">", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp > 0 {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put(">=", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			cmp, err := compareNumber(parameters, ">=", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp >= 0 {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("<", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			cmp, err := compareNumber(parameters, "<", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp < 0 {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("<=", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			cmp, err := compareNumber(parameters, "<=", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp <= 0 {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("=", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			cmp, err := compareNumber(parameters, "=", evaluator, environment)
			if err != nil {
				return nil, err
			}
			if cmp == 0 {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("and", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			result := true
			for _, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				boolVal, ok := val.(ConstantValue)
				if !ok {
					return nil, fmt.Errorf("expected boolean value, got %T", val)
				}
				if boolVal == FalseValue {
					result = false
					break
				}
			}
			if result {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("or", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			result := false
			for _, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				boolVal, ok := val.(ConstantValue)
				if !ok {
					return nil, fmt.Errorf("expected boolean value, got %T", val)
				}
				if boolVal == TrueValue {
					result = true
					break
				}
			}
			if result {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	// cons
	env.Put("cons", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 2 {
				return nil, fmt.Errorf("'cons' has been called with %d arguments; it requires exactly 2 argument", len(parameters))
			}

			car, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}

			cdr, err := evaluator.eval(parameters[1], environment)
			if err != nil {
				return nil, err
			}
			res := &ConsValue{
				Car: car,
				Cdr: cdr,
			}
			return res, nil
		},
	})

	env.Put("list", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			elements := make([]ReturnValue, len(parameters))
			for i, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				elements[i] = val
			}
			return &ListValue{Elements: elements}, nil
		},
	})

	env.Put("append", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) < 2 {
				return nil, fmt.Errorf("`append` has been called with %d arguments; it requires at lesat 2 argument", len(parameters))
			}
			elements := make([]ReturnValue, 0)
			for _, parameter := range parameters {
				val, err := evaluator.eval(parameter, environment)
				if err != nil {
					return nil, err
				}
				listVal, ok := val.(*ListValue)
				if !ok {
					return nil, fmt.Errorf("expected list value, got %T", val)
				}
				elements = append(elements, listVal.Elements...)
			}
			return &ListValue{Elements: elements}, nil
		},
	})

	env.Put("car", ConProcedureFactory([]ConOperation{CON_OP_CAR}))
	env.Put("cdr", ConProcedureFactory([]ConOperation{CON_OP_CDR}))
	env.Put("cadr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CAR}))
	env.Put("cdar", ConProcedureFactory([]ConOperation{CON_OP_CAR, CON_OP_CDR}))
	env.Put("caddr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CDR, CON_OP_CAR}))
	env.Put("cadddr", ConProcedureFactory([]ConOperation{CON_OP_CDR, CON_OP_CDR, CON_OP_CDR, CON_OP_CAR}))

	env.Put("null?", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'null?' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			listVal, ok := val.(*ListValue)
			if !ok {
				return FalseValue, nil
			}
			if len(listVal.Elements) == 0 {
				return TrueValue, nil
			} else {
				return FalseValue, nil
			}
		},
	})

	env.Put("display", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'display' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			if str, ok := val.(*StringValue); ok {
				fmt.Print(str.Value)
			} else {
				fmt.Print(val.String())
			}

			return VoidConst, nil
		},
	})

	// TODO: how to print stack trace on error?

	env.Put("newline", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 0 {
				return nil, fmt.Errorf("'newline' has been called with %d arguments; it requires exactly 0 argument", len(parameters))
			}

			fmt.Println()

			return VoidConst, nil
		},
	})

	env.Put("print", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("'print' has been called with %d arguments; it requires exactly 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			fmt.Println(val.String())

			return VoidConst, nil
		},
	})

	env.Put("error", &BuiltinFunction{
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) < 1 {
				return nil, fmt.Errorf("'error' has been called with %d arguments; it requires at least 1 argument", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}
			//fmt.Println(val.String())

			panic(fmt.Sprintf("failed to evaluate: %s", val.String()))

			//return VoidConst, nil
		},
	})

	// Add more built-in functions as needed
	return env
}

func (e *Evaluator) evalNumber(parameter parser.Expression, environment *Environment) (*NumberValue, error) {
	val, err := e.eval(parameter, environment)
	if err != nil {
		return nil, err
	}
	numVal, ok := val.(*NumberValue)
	if !ok {
		return nil, fmt.Errorf("expected number value, got %T", val)
	}
	return numVal, nil
}

func compareNumber(parameters []parser.Expression, op string, evaluator *Evaluator, environment *Environment) (int, error) {
	if len(parameters) != 2 {
		return 0, fmt.Errorf("'%s' has been called with %d arguments; it requires exactly 1 argument", op, len(parameters))
	}

	left, err := evaluator.eval(parameters[0], environment)
	if err != nil {
		return 0, err
	}
	leftVal, ok := left.(*NumberValue)
	if !ok {
		return 0, fmt.Errorf("expected number value, got %T", leftVal)
	}

	right, err := evaluator.eval(parameters[1], environment)
	if err != nil {
		return 0, err
	}
	rightVal, ok := right.(*NumberValue)
	if !ok {
		return 0, fmt.Errorf("expected number value, got %T", rightVal)
	}

	if leftVal.Value > rightVal.Value {
		return 1, nil
	} else if leftVal.Value < rightVal.Value {
		return -1, nil
	} else {
		return 0, nil
	}
}

type Environment struct {
	enclosing *Environment
	store     map[string]ReturnValue
}

func newEnvironment() *Environment {
	return &Environment{
		store: make(map[string]ReturnValue),
	}
}

func (env *Environment) Put(key string, value ReturnValue) {
	env.store[key] = value
}

func (env *Environment) Get(key string) (ReturnValue, bool) {
	val, ok := env.store[key]
	if !ok && env.enclosing != nil {
		return env.enclosing.Get(key)
	}
	return val, ok
}

// Update updates the value of an existing key in the environment and returns the old value.
// If the key does not exist in the current environment, it recursively
// checks the enclosing environment. If the key is not found in any
// environment, it returns an error.
func (env *Environment) Update(key string, value ReturnValue) (ReturnValue, error) {
	oldVal, ok := env.store[key]
	if ok {
		env.store[key] = value
		return oldVal, nil
	} else if env.enclosing != nil {
		return env.enclosing.Update(key, value)
	}

	return nil, fmt.Errorf("can't find key %s to update", key)
}

func (e *Evaluator) Eval(program *parser.Program) (ReturnValue, error) {
	var ret ReturnValue
	var err error
	for _, exp := range program.Expressions {
		ret, err = e.eval(exp, e.globalEnv)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (e *Evaluator) eval(expression parser.Expression, environment *Environment) (ReturnValue, error) {
	switch expression {
	case parser.TrueLiteral:
		return TrueValue, nil
	case parser.FalseLiteral:
		return FalseValue, nil
	case parser.Void:
		return VoidConst, nil
	}

	switch exp := expression.(type) {
	case *parser.NumberLiteral:
		return &NumberValue{Value: exp.Value}, nil
	case *parser.StringLiteral:
		return &StringValue{Value: exp.Value}, nil
	case *parser.SymbolExpression:
		return &SymbolValue{Value: exp.Value}, nil
	case *parser.DefineExpression:
		return e.evalDefineExpression(exp, environment)
	case *parser.IdentifierExpression:
		val, ok := environment.Get(exp.Value)
		if !ok {
			return nil, fmt.Errorf("undefined identifier: `%s` on line %d", exp.Value, exp.Token().Line)
		}
		return val, nil
	case *parser.CallExpression:
		return e.evalCallExpression(exp, environment)
	case *parser.LambdaExpression:
		return e.evalLambdaExpression(exp, environment)
	case *parser.PrimitiveProcedureExpression:
		fn, ok := environment.Get(exp.String())
		if !ok {
			panic("undefined primitive identifier: `" + exp.String() + "`")
		}
		builtinFn, ok := fn.(*BuiltinFunction)
		return builtinFn, nil
	case *parser.IfExpression:
		return e.evalIfExpression(exp, environment)
	case *parser.SetExpression:
		return e.evalSetExpression(exp, environment)
	case *parser.ListExpression:
		return e.evalListExpression(exp, environment)
	default:

		return nil, fmt.Errorf("unsupported expression type: %T", exp)
	}
}

func (e *Evaluator) evalListExpression(exp *parser.ListExpression, environment *Environment) (ReturnValue, error) {
	elements := make([]ReturnValue, len(exp.Elements))
	for i, element := range exp.Elements {
		val, err := e.eval(element, environment)
		if err != nil {
			return nil, err
		}
		elements[i] = val
	}
	return &ListValue{Elements: elements}, nil
}

func (e *Evaluator) evalSetExpression(exp *parser.SetExpression, environment *Environment) (ReturnValue, error) {
	val, err := e.eval(exp.Value, environment)
	if err != nil {
		return nil, err
	}

	return environment.Update(exp.Name, val)
}

func (e *Evaluator) evalIfExpression(exp *parser.IfExpression, environment *Environment) (ReturnValue, error) {
	cond, err := e.eval(exp.Predicate, environment)
	if err != nil {
		return nil, err
	}

	condVal, ok := cond.(ConstantValue)
	if !ok {
		return nil, fmt.Errorf("condition must evaluate to a bool, got %T", condVal)
	}
	if condVal == TrueValue {
		return e.eval(exp.Consequent, environment)
	} else if condVal == FalseValue {
		if exp.Alternative != nil {
			return e.eval(exp.Alternative, environment)
		} else {
			return VoidConst, nil
		}
	} else {
		return nil, fmt.Errorf("condition must evaluate to a bool, got %v", condVal)
	}
}

func (e *Evaluator) evalLambdaExpression(exp *parser.LambdaExpression, environment *Environment) (ReturnValue, error) {
	params := make([]string, len(exp.Parameters))
	for i, param := range exp.Parameters {
		params[i] = param
	}

	proc := &ProcedureValue{
		Parameters:            params,
		OptionalTailParameter: exp.OptionalTailParameter,
		Body:                  exp.Body,
		Env:                   environment,
	}
	return proc, nil
}

func (e *Evaluator) evalDefineExpression(exp *parser.DefineExpression, environment *Environment) (ReturnValue, error) {
	val, err := e.eval(exp.Value, environment)
	if err != nil {
		return nil, err
	}
	environment.Put(exp.Name, val)
	return val, nil
}

// how to handle runtime error with stack trace?
type RuntimeError struct {
	rawErrorMessage string
	Stack           []lexer.Token
}

func (e *RuntimeError) Error() string {
	return e.rawErrorMessage
}

func newRuntimeError(err error, token lexer.Token) *RuntimeError {
	var prevError *RuntimeError
	if ok := errors.As(err, &prevError); ok {
		stacks := make([]lexer.Token, 0, len(prevError.Stack)+1)
		stacks = append(stacks, token)
		stacks = append(stacks, prevError.Stack...)
		return &RuntimeError{rawErrorMessage: prevError.rawErrorMessage, Stack: stacks}
	} else {
		return &RuntimeError{rawErrorMessage: err.Error(), Stack: []lexer.Token{token}}
	}
}

func (e *Evaluator) evalCallExpression(exp *parser.CallExpression, environment *Environment) (ReturnValue, error) {
	operator := exp.Operator
	val, err := e.eval(operator, environment)
	if err != nil {
		return nil, err
	}
	switch fn := val.(type) {
	case *BuiltinFunction:
		return e.evalBuiltinFunction(fn, exp.Operands, environment)
	case *ProcedureValue:
		ret, err := e.evalProcedure(fn, exp.Operands, environment)
		if err != nil {
			return nil, newRuntimeError(err, operator.Token())
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("unsupported operator type: %T", fn)
	}
}

func (e *Evaluator) evalBuiltinFunction(builtinFn *BuiltinFunction, operands []parser.Expression, environment *Environment) (ReturnValue, error) {
	ret, err := builtinFn.Fn(operands, e, environment)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (e *Evaluator) evalProcedure(procedure *ProcedureValue, operands []parser.Expression, environment *Environment) (ReturnValue, error) {
	if procedure.CaneTakeArbitraryParameters() {
		if len(procedure.Parameters) > len(operands) {
			return nil, fmt.Errorf("expected at least %d arguments, got %d", len(procedure.Parameters), len(operands))
		}
	} else if len(procedure.Parameters) != len(operands) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(procedure.Parameters), len(operands))
	}

	// Create a new environment for the procedure call
	newEnv := newEnvironment()
	newEnv.enclosing = procedure.Env

	// Evaluate arguments and bind them to parameters in the new environment
	for i, param := range procedure.Parameters {
		argVal, err := e.eval(operands[i], environment)
		if err != nil {
			return nil, err
		}
		newEnv.Put(param, argVal)
	}

	if procedure.CaneTakeArbitraryParameters() {
		tailArgs := ListValue{Elements: make([]ReturnValue, 0)}
		for i := len(procedure.Parameters); i < len(operands); i++ {
			argVal, err := e.eval(operands[i], environment)
			if err != nil {
				return nil, err
			}
			tailArgs.Elements = append(tailArgs.Elements, argVal)
		}

		newEnv.Put(procedure.OptionalTailParameter, &tailArgs)
	}

	// Evaluate the body of the procedure in the new environment
	var result ReturnValue
	var err error
	for _, expr := range procedure.Body {
		result, err = e.eval(expr, newEnv)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
