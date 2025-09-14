package evaluator

import (
	"fmt"

	"github.com/ocowchun/soup/parser"
)

func car(val ReturnValue) (ReturnValue, error) {
	switch val := val.(type) {
	case *ConsValue:
		return val.Car, nil
	case *ListValue:
		if len(val.Elements) == 0 {
			return nil, fmt.Errorf("cannot call 'car' on an empty list")
		}
		return val.Elements[0], nil
	default:
		return nil, fmt.Errorf("expected cons or list value, got %T", val)
	}
}
func cdr(val ReturnValue) (ReturnValue, error) {
	switch val := val.(type) {
	case *ConsValue:
		return val.Cdr, nil
	case *ListValue:
		if len(val.Elements) == 0 {
			return nil, fmt.Errorf("cannot call 'cdr' on an empty list")
		}
		return &ListValue{Elements: val.Elements[1:]}, nil
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
		Fn: func(parameters []parser.Expression, evaluator *Evaluator, environment *Environment) (ReturnValue, error) {
			if len(parameters) != 1 {
				return nil, fmt.Errorf("expected 1 argument, got %d", len(parameters))
			}

			val, err := evaluator.eval(parameters[0], environment)
			if err != nil {
				return nil, err
			}

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

//fnFactory := func (operations []str) {
//
//}
