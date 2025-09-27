package evaluator

import "fmt"

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
