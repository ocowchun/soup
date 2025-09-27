package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ocowchun/soup/lexer"
)

type ValueType uint8

const (
	NumberType ValueType = iota
	StringType
	BooleanType
)

func (t ValueType) String() string {
	switch t {
	case NumberType:
		return "Number"
	case StringType:
		return "String"
	case BooleanType:
		return "Boolean"
	default:
		return "Unknown"
	}

}

type Value struct {
	Type ValueType
	Data any
}

type Environment struct {
	enclosing *Environment
	values    map[string]*Value
}

func NewEnvironment() *Environment {
	return &Environment{
		enclosing: nil,
		values:    make(map[string]*Value),
	}
}

func (env *Environment) Get(name string) (*Value, bool) {
	val, exists := env.values[name]
	if !exists && env.enclosing != nil {
		return env.enclosing.Get(name)
	}
	return val, exists
}

func (env *Environment) Define(name string, value *Value) {
	env.values[name] = value
}
func (env *Environment) Update(name string, value *Value) bool {
	if _, exists := env.values[name]; exists {
		env.values[name] = value
		return true
	}
	if env.enclosing != nil {
		return env.enclosing.Update(name, value)
	}
	return false
}

type Element interface {
	element()
}
type List struct {
	Elements []Element
}

func (l *List) element() {}

type MyElement struct {
	Token lexer.Token
}

func (l *MyElement) element() {}

func readList(l *lexer.Lexer) (*List, error) {
	list := &List{Elements: make([]Element, 0)}
	for {
		token := l.NextToken()
		if token.TokenType == lexer.TokenTypeRightParen {
			return list, nil
		} else if token.TokenType == lexer.TokenTypeLeftParen {
			element, err := readList(l)
			if err != nil {
				return nil, err
			}
			list.Elements = append(list.Elements, element)
		} else {
			element := &MyElement{Token: token}
			list.Elements = append(list.Elements, element)
		}
	}

}

func main() {
	reader := bufio.NewReader(os.Stdin)
	//text, _ := reader.ReadString('\n')
	//fmt.Print("You entered: ", text)
	l := lexer.New(reader)

	firstTok := l.NextToken()
	if firstTok.TokenType == lexer.TokenTypeLeftParen {

	} else {
		fmt.Println(firstTok.Content)
	}
	fmt.Println("done")
	//p := parser.New(l)
	//program, err := p.Parse()
	//if err != nil {
	//	panic(err)
	//}
	//
	//for _, exp := range program.Expressions {
	//	fmt.Println(exp)
	//}

	//fmt.Println("Hello World")
	//v := &Value{Type: NumberType, Data: 42}
	//env := NewEnvironment()
	//env.Define("answer", v)
	//
	//foo := 1 | 2
	//fmt.Println(foo)

	//v.Data = "123"
	//v.Type = StringType
	//v.Data = 123
	//if num, ok := v.Data.(int); ok {
	//	fmt.Println("int", num)
	//}
	//if num, ok := v.Data.(float64); ok {
	//	fmt.Println("float64", num)
	//}
	//v.Data = float64(235)
	//if num, ok := v.Data.(int); ok {
	//	fmt.Println("int", num)
	//}
	//if num, ok := v.Data.(float64); ok {
	//	fmt.Println("float64", num)
	//}
	//
	//val, exists := env.Get("answer")
	//if exists {
	//	fmt.Printf("Value: %v, Type: %v\n", val.Data, val.Type)
	//} else {
	//	fmt.Println("Value not found")
	//}
	//
	//{
	//	text := "1.2"
	//	num, err := strconv.ParseInt(text, 10, 64)
	//	if err != nil {
	//		fmt.Println("err", err)
	//	} else {
	//		fmt.Println("int64", num)
	//	}
	//
	//	f, err := strconv.ParseFloat(text, 64)
	//	if err != nil {
	//		fmt.Println("err", err)
	//	} else {
	//		fmt.Println("float64", f)
	//	}
	//
	//}

}
