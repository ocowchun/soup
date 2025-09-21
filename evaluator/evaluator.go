package evaluator

import (
	"fmt"
	"strconv"

	"github.com/ocowchun/soup/parser"
)

type Evaluator struct {
	globalEnv      *Environment
	procedureNames []string
}

func New() *Evaluator {
	env := initGlobalEnvironment()
	return &Evaluator{globalEnv: env, procedureNames: []string{}}
}

func (e *Evaluator) currentProcedureName() string {
	return e.procedureNames[len(e.procedureNames)-1]
}

func (e *Evaluator) pushProcedureName(newProcedureName string) {
	e.procedureNames = append(e.procedureNames, newProcedureName)
}

func (e *Evaluator) popProcedureName() string {
	res := e.procedureNames[len(e.procedureNames)-1]
	e.procedureNames = e.procedureNames[:len(e.procedureNames)-1]
	return res
}

func equal(a *ReturnValue, b *ReturnValue) bool {
	if a == b {
		return true
	}
	switch a.Type {
	case NumberType:
		if b.Type != NumberType {
			return false
		}
		return a.Number() == b.Number()
	case StringType:
		if b.Type != StringType {
			return false
		}
		return a.String() == b.String()
	case SymbolType:
		if b.Type != SymbolType {
			return false
		}
		return a.String() == b.String()
	case ListType:
		if b.Type != ListType {
			return false
		}
		aList := a.List()
		bList := b.List()
		if len(aList.Elements) != len(bList.Elements) {
			return false
		}
		for i := range aList.Elements {
			if !equal(aList.Elements[i], bList.Elements[i]) {
				return false
			}
		}
		return true
	case ConstantType:
		if b.Type != ConstantType {
			return false
		}
		return a.Constant() == b.Constant()
	case ConsType:
		if b.Type != ConsType {
			return false
		}
		aCons := a.Cons()
		bCons := b.Cons()
		return equal(aCons.Car, bCons.Car) && equal(aCons.Cdr, bCons.Cdr)
	default:
		return false
	}
}

func addBuiltinToEnv(env *Environment, name string, fn *BuiltinFunction) {
	env.Put(name, &ReturnValue{Type: BuiltinFunctionType, Data: fn})
}

type Number struct {
	// bad idea, try to improve it later
	data any
}

func MakeFloat64Number(data float64) Number {
	return Number{data: data}
}

func MakeInt64Number(data int64) Number {
	return Number{data: data}
}

func (n Number) isInt64() bool {
	_, ok := n.data.(int64)
	return ok
}

func (n Number) Int64() int64 {
	num, ok := n.data.(int64)
	if !ok {
		panic("number is not int64")
	}
	return num
}

func (n Number) Float64() float64 {
	num, ok := n.data.(int64)
	if ok {
		return float64(num)
	}
	return n.data.(float64)
}

func (n Number) String() string {
	if num, ok := n.data.(int64); ok {
		return fmt.Sprintf("%v", num)
	}
	return fmt.Sprintf("%v", n.data.(float64))
}

type Environment struct {
	enclosing *Environment
	store     map[string]*ReturnValue
}

func newEnvironment() *Environment {
	return &Environment{
		store: make(map[string]*ReturnValue),
	}
}

func (env *Environment) Put(key string, value *ReturnValue) {
	env.store[key] = value
}

func (env *Environment) Get(key string) (*ReturnValue, bool) {
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
func (env *Environment) Update(key string, value *ReturnValue) (*ReturnValue, error) {
	oldVal, ok := env.store[key]
	if ok {
		env.store[key] = value
		return oldVal, nil
	} else if env.enclosing != nil {
		return env.enclosing.Update(key, value)
	}

	return nil, fmt.Errorf("can't find key %s to update", key)
}

func (e *Evaluator) Eval(program *parser.Program) (*ReturnValue, error) {
	var ret *ReturnValue
	var err error
	e.procedureNames = append(e.procedureNames, "main")
	for _, exp := range program.Expressions {
		ret, err = e.eval(exp, e.globalEnv)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func (e *Evaluator) eval(expression parser.Expression, environment *Environment) (*ReturnValue, error) {
	switch expression {
	case parser.TrueLiteral:
		return &ReturnValue{Type: ConstantType, Data: TrueValue}, nil
	case parser.FalseLiteral:
		return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
	case parser.Void:
		return &ReturnValue{Type: ConstantType, Data: VoidConst}, nil
	}

	switch exp := expression.(type) {
	case *parser.NumberLiteral:
		if data, err := strconv.ParseInt(exp.NumToken.Content, 10, 64); err == nil {
			return &ReturnValue{Type: NumberType, Data: Number{data: data}}, nil
		}

		f, err := strconv.ParseFloat(exp.NumToken.Content, 64)
		if err != nil {
			panic(err)
		}
		return &ReturnValue{Type: NumberType, Data: Number{data: f}}, nil
	case *parser.StringLiteral:
		return &ReturnValue{Type: StringType, Data: exp.Value}, nil
	case *parser.SymbolExpression:
		return &ReturnValue{Type: SymbolType, Data: exp.Value}, nil
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
		if fn.Type != BuiltinFunctionType {
			return nil, fmt.Errorf("identifier `%s` is not a builtin function", exp.String())
		}
		return fn, nil
	case *parser.IfExpression:
		return e.evalIfExpression(exp, environment)
	case *parser.SetExpression:
		return e.evalSetExpression(exp, environment)
	case *parser.ListExpression:
		return e.evalListExpression(exp, environment)
	case *parser.BeginExpression:
		return e.evalBeginExpression(exp, environment)
	default:

		return nil, fmt.Errorf("unsupported expression type: %T", exp)
	}
}

func (e *Evaluator) evalBeginExpression(exp *parser.BeginExpression, environment *Environment) (*ReturnValue, error) {
	for i, subExp := range exp.Expressions {
		val, err := e.eval(subExp, environment)
		if err != nil {
			return nil, err
		}

		if i == len(exp.Expressions)-1 {
			return val, nil
		}
	}
	panic("unreachable")
}

func (e *Evaluator) evalListExpression(exp *parser.ListExpression, environment *Environment) (*ReturnValue, error) {
	elements := make([]*ReturnValue, len(exp.Elements))
	for i, element := range exp.Elements {
		val, err := e.eval(element, environment)
		if err != nil {
			return nil, err
		}
		elements[i] = val
	}
	list := &ListValue{Elements: elements}
	return &ReturnValue{Type: ListType, Data: list}, nil
}

func (e *Evaluator) evalSetExpression(exp *parser.SetExpression, environment *Environment) (*ReturnValue, error) {
	val, err := e.eval(exp.Value, environment)
	if err != nil {
		return nil, err
	}

	return environment.Update(exp.Name, val)
}

func (e *Evaluator) evalIfExpression(exp *parser.IfExpression, environment *Environment) (*ReturnValue, error) {
	cond, err := e.eval(exp.Predicate, environment)
	if err != nil {
		return nil, newRuntimeError(err, exp.Predicate.Token(), e.currentProcedureName())
	}

	// In Scheme, any value except #f counts as true in conditionals.
	// https://docs.scheme.org/schintro/schintro_87.html
	if cond.Type == ConstantType && cond.Data == FalseValue {
		if exp.Alternative != nil {
			ret, err := e.eval(exp.Alternative, environment)
			if err != nil {
				return nil, newRuntimeError(err, exp.Alternative.Token(), e.currentProcedureName())
				//return nil, err
			}
			return ret, nil
		} else {
			return &ReturnValue{Type: ConstantType, Data: VoidConst}, nil
		}
	} else {
		ret, err := e.eval(exp.Consequent, environment)
		if err != nil {
			return nil, newRuntimeError(err, exp.Consequent.Token(), e.currentProcedureName())
		}
		return ret, nil
	}
}

func (e *Evaluator) evalLambdaExpression(exp *parser.LambdaExpression, environment *Environment) (*ReturnValue, error) {
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
	return &ReturnValue{Type: ProcedureType, Data: proc}, nil
}

func (e *Evaluator) evalDefineExpression(exp *parser.DefineExpression, environment *Environment) (*ReturnValue, error) {
	val, err := e.eval(exp.Value, environment)
	if err != nil {
		return nil, err
	}
	environment.Put(exp.Name, val)
	return val, nil
}

func (e *Evaluator) evalCallExpression(exp *parser.CallExpression, environment *Environment) (*ReturnValue, error) {
	operator := exp.Operator

	val, err := e.eval(operator, environment)
	if err != nil {
		return nil, newRuntimeError(err, operator.Token(), e.currentProcedureName())
	}

	isOrFn := val.Type == BuiltinFunctionType && operator.String() == "or"
	isAndFn := val.Type == BuiltinFunctionType && operator.String() == "and"

	operands := make([]*ReturnValue, len(exp.Operands))
	for i, op := range exp.Operands {
		operand, err := e.eval(op, environment)
		if err != nil {
			return nil, newRuntimeError(err, operator.Token(), e.currentProcedureName())
		}
		// Workaround to support (or 1 bad-exp), to not eval bad-exp
		if isOrFn && !(operand.Type == ConstantType && operand.Data == FalseValue) {
			return operand, nil
		}

		// Workaround to support (and #f bad-exp), to not eval bad-exp
		if isAndFn && (operand.Type == ConstantType && operand.Data == FalseValue) {
			return &ReturnValue{Type: ConstantType, Data: FalseValue}, nil
		}

		operands[i] = operand
	}

	switch val.Type {
	case BuiltinFunctionType:
		e.pushProcedureName(operator.String())

		fn := val.BuiltinFunction()
		ret, err := e.evalBuiltinFunction(fn, operands, environment)
		if err != nil {
			fmt.Println("error", operator.String(), err)
			return nil, newRuntimeError(err, operator.Token(), e.popProcedureName())
		}

		e.popProcedureName()
		return ret, nil

	case ProcedureType:
		e.pushProcedureName(operator.String())
		fn := val.Procedure()
		ret, err := e.evalProcedure(fn, operands, environment)
		if err != nil {
			return nil, newRuntimeError(err, operator.Token(), e.popProcedureName())
		}
		e.popProcedureName()
		return ret, nil
	default:
		err = fmt.Errorf("unsupported operator type: %s(%s)", val.Type, val.String())
		return nil, newRuntimeError(err, operator.Token(), e.currentProcedureName())
	}
}

func (e *Evaluator) evalBuiltinFunction(builtinFn *BuiltinFunction, operands []*ReturnValue, environment *Environment) (*ReturnValue, error) {
	ret, err := builtinFn.Fn(operands, e, environment)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (e *Evaluator) evalProcedure(procedure *ProcedureValue, operands []*ReturnValue, environment *Environment) (*ReturnValue, error) {
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
		newEnv.Put(param, operands[i])
	}

	if procedure.CaneTakeArbitraryParameters() {
		tailArgs := ListValue{Elements: make([]*ReturnValue, 0)}
		for i := len(procedure.Parameters); i < len(operands); i++ {
			tailArgs.Elements = append(tailArgs.Elements, operands[i])
		}

		newEnv.Put(procedure.OptionalTailParameter, &ReturnValue{Type: ListType, Data: &tailArgs})
	}

	// Evaluate the body of the procedure in the new environment
	var result *ReturnValue
	var err error
	for _, expr := range procedure.Body {
		result, err = e.eval(expr, newEnv)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
