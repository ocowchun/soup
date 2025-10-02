package parser

import (
	"fmt"
	"strconv"

	"github.com/ocowchun/soup/lexer"
)

// we might not need a parser for lisp-like language??
// we still need parser, because the program might contains multiple expressions
type Parser struct {
	l            *lexer.Lexer
	prevToken    lexer.Token
	currentToken lexer.Token
}

func (p *Parser) nextToken() {
	token := p.l.NextToken()
	p.prevToken = p.currentToken
	p.currentToken = token
}

type Program struct {
	Expressions []Expression
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{l: l}
}

func (p *Parser) match(t lexer.TokenType) bool {
	if p.currentToken.TokenType == t {
		p.nextToken()
		return true
	}

	return false
}

func (p *Parser) Parse() (*Program, error) {
	program := &Program{Expressions: []Expression{}}

	p.nextToken()

	for {
		if p.match(lexer.TokenTypeEOF) {
			break
		}

		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		program.Expressions = append(program.Expressions, expr)

	}
	return program, nil
}

type ParsingError struct {
	Message string
	Token   lexer.Token
}

func (e *ParsingError) Error() string {
	return e.Message
}

func NewParsingError(token lexer.Token, message string) *ParsingError {
	return &ParsingError{
		Message: message,
		Token:   token,
	}
}

func (p *Parser) parseNumber() (*NumberLiteral, error) {
	_, err := strconv.ParseFloat(p.currentToken.Content, 64)
	if err != nil {
		return nil, NewParsingError(p.currentToken, err.Error())
	}

	exp := &NumberLiteral{
		NumToken: p.currentToken,
	}
	p.nextToken()

	return exp, err
}

func (p *Parser) parseString() (Expression, error) {
	str := &StringLiteral{Value: p.currentToken.Content, StrToken: p.currentToken}
	p.nextToken()
	return str, nil
}

func (p *Parser) parseCallExpression() (Expression, error) {
	currentToken := p.currentToken
	operator, err := p.parseExpression()
	if err != nil {
		return nil, NewParsingError(currentToken, err.Error())
	}

	operands := make([]Expression, 0)
	for p.currentToken.TokenType != lexer.TokenTypeRightParen {
		currentToken = p.currentToken
		operand, err := p.parseExpression()
		if err != nil {
			return nil, NewParsingError(currentToken, err.Error())
		}
		operands = append(operands, operand)
	}

	p.nextToken()

	exp := &CallExpression{
		LeftParenToken: currentToken,
		Operator:       operator,
		Operands:       operands,
	}

	return exp, nil
}

func (p *Parser) parseIfExpression() (Expression, error) {
	firstToken := p.currentToken
	p.nextToken()

	predicate, err := p.parseExpression()
	if err != nil {
		return nil, NewParsingError(p.currentToken, err.Error())
	}

	consequent, err := p.parseExpression()
	if err != nil {
		return nil, NewParsingError(p.currentToken, err.Error())
	}

	var alternative Expression = Void
	if p.currentToken.TokenType != lexer.TokenTypeRightParen {
		alternative, err = p.parseExpression()
		if err != nil {
			return nil, NewParsingError(p.currentToken, err.Error())
		}
		if p.currentToken.TokenType != lexer.TokenTypeRightParen {
			return nil, NewParsingError(p.currentToken, "expected ')' after if expression")
		}
	}

	p.nextToken()

	return &IfExpression{
		LeftParenToken: firstToken,
		Predicate:      predicate,
		Consequent:     consequent,
		Alternative:    alternative,
	}, nil
}

func (p *Parser) parseDefineExpression() (Expression, error) {
	firstToken := p.currentToken
	p.nextToken()

	if p.currentToken.TokenType == lexer.TokenTypeLeftParen {
		// TODO: handle dotted-tail notation
		// In a procedure definition, a parameter list that has a dot before the last parameter name indicates that,
		// when the procedure is called, the initial parameters (if any) will have as values the initial arguments,
		// as usual, but the final parameter’s value will be a list of any remaining arguments. For instance,
		// given the definition
		//(define (f x y . z) ⟨body⟩)
		//the procedure f can be called with two or more arguments. If we evaluate
		//
		//(f 1 2 3 4 5 6)
		//then in the body of f, x will be 1, y will be 2, and z will be the list (3 4 5 6)
		// (define (name params...) body...)
		p.nextToken()

		if p.currentToken.TokenType != lexer.TokenTypeIdentifier {
			return nil, NewParsingError(p.currentToken, "expected identifier after '(' in define")
		}
		name := p.currentToken.Content

		p.nextToken()

		parameters := make([]string, 0)
		optionalTailParameter := ""
		for p.currentToken.TokenType != lexer.TokenTypeRightParen {
			// parse parameters
			// TODO: how to adjust struct to support dotted-tail notation?
			if p.currentToken.TokenType == lexer.TokenTypeDot {
				p.nextToken()
				if p.currentToken.TokenType != lexer.TokenTypeIdentifier {
					return nil, NewParsingError(p.currentToken, "expected identifier in parameter list")
				}

				optionalTailParameter = p.currentToken.Content
				p.nextToken()
				if p.currentToken.TokenType == lexer.TokenTypeRightParen {
					break
				} else {
					return nil, NewParsingError(p.currentToken, "expected ')' after optional tail parameter")
				}
			}

			if p.currentToken.TokenType != lexer.TokenTypeIdentifier {
				return nil, NewParsingError(p.currentToken, "expected identifier in parameter list")
			}
			parameters = append(parameters, p.currentToken.Content)

			p.nextToken()
		}
		p.nextToken()

		body := make([]Expression, 0)
		for p.currentToken.TokenType != lexer.TokenTypeRightParen {
			expr, err := p.parseExpression()
			if err != nil {
				return nil, NewParsingError(p.currentToken, err.Error())
			}
			body = append(body, expr)
		}
		if len(body) == 0 {
			return nil, NewParsingError(p.currentToken, "expected at least one expression in function body")
		}

		p.nextToken()

		lambda := &LambdaExpression{
			Parameters:            parameters,
			Body:                  body,
			OptionalTailParameter: optionalTailParameter,
		}
		return &DefineExpression{
			LeftParenToken: firstToken,
			Name:           name,
			Value:          lambda,
		}, nil
	} else {
		// (define name body...) -> variable
		if p.currentToken.TokenType != lexer.TokenTypeIdentifier {
			return nil, NewParsingError(p.currentToken, "expected identifier after define")
		}
		name := p.currentToken.Content
		p.nextToken()

		exp, err := p.parseExpression()
		if err != nil {
			return nil, NewParsingError(p.currentToken, err.Error())
		}

		if p.currentToken.TokenType != lexer.TokenTypeRightParen {
			return nil, NewParsingError(p.currentToken, "expected ')' after define expression")
		}

		p.nextToken()
		return &DefineExpression{
			LeftParenToken: firstToken,
			Name:           name,
			Value:          exp,
		}, nil
	}
}

func (p *Parser) parseLambdaExpression() (Expression, error) {
	firstToken := p.currentToken
	p.nextToken()

	// (lambda (params...) body...)
	if p.currentToken.TokenType != lexer.TokenTypeLeftParen {
		return nil, NewParsingError(p.currentToken, "expected '(' after lambda")
	}

	p.nextToken()

	parameters := make([]string, 0)
	for p.currentToken.TokenType != lexer.TokenTypeRightParen {
		if p.currentToken.TokenType != lexer.TokenTypeIdentifier {
			return nil, NewParsingError(p.currentToken, "expected identifier in parameter list")
		}
		parameters = append(parameters, p.currentToken.Content)

		p.nextToken()
	}

	p.nextToken()

	body := make([]Expression, 0)
	for p.currentToken.TokenType != lexer.TokenTypeRightParen {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, NewParsingError(p.currentToken, err.Error())
		}
		body = append(body, expr)
	}

	if len(body) == 0 {
		return nil, NewParsingError(p.currentToken, "expected at least one expression in lambda body")
	}

	p.nextToken()
	return &LambdaExpression{
		LeftParenToken: firstToken,
		Parameters:     parameters,
		Body:           body,
	}, nil
}

func (p *Parser) parseLetExpression() (Expression, error) {
	firstToken := p.currentToken
	p.nextToken()

	// (let ( (var expr) ... ) body ... )
	if p.currentToken.TokenType != lexer.TokenTypeLeftParen {
		return nil, NewParsingError(p.currentToken, "expected '(' after let")
	}

	p.nextToken()

	parameterNames := make([]string, 0)
	parameterExprs := make([]Expression, 0)
	for p.currentToken.TokenType != lexer.TokenTypeRightParen {
		if p.currentToken.TokenType != lexer.TokenTypeLeftParen {
			return nil, NewParsingError(p.currentToken, "expected '(' in binding list")
		}
		p.nextToken()

		if p.currentToken.TokenType != lexer.TokenTypeIdentifier {
			return nil, NewParsingError(p.currentToken, "expected identifier in binding")
		}
		parameterName := p.currentToken.Content

		p.nextToken()

		parameterExp, err := p.parseExpression()
		if err != nil {
			return nil, NewParsingError(p.currentToken, err.Error())
		}

		if p.currentToken.TokenType != lexer.TokenTypeRightParen {
			return nil, NewParsingError(p.currentToken, "expected ')' after binding")
		}

		p.nextToken()
		// TODO: check duplicate parameter names
		parameterNames = append(parameterNames, parameterName)
		parameterExprs = append(parameterExprs, parameterExp)
	}

	p.nextToken()

	body := make([]Expression, 0)
	for p.currentToken.TokenType != lexer.TokenTypeRightParen {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, NewParsingError(p.currentToken, err.Error())
		}
		body = append(body, expr)
	}
	if len(body) == 0 {
		return nil, NewParsingError(p.currentToken, "expected at least one expression in let body")
	}
	p.nextToken()

	lambda := &LambdaExpression{
		LeftParenToken: firstToken,
		Parameters:     parameterNames,
		Body:           body,
	}
	return &CallExpression{
		LeftParenToken: firstToken,
		Operator:       lambda,
		Operands:       parameterExprs,
	}, nil
}

func (p *Parser) parseCondExpression() (Expression, error) {
	//return nil, fmt.Errorf("not implemented")

	// (cond
	//((predicate1) exp)
	//((predicate2) exp)
	//(else exp)
	//)
	ifFirstToken := p.currentToken

	p.nextToken()
	var ifExp *IfExpression
	var currentIfExp = ifExp
	for {
		if p.currentToken.TokenType == lexer.TokenTypeRightParen {
			break
		}
		if !p.match(lexer.TokenTypeLeftParen) {
			return nil, NewParsingError(p.currentToken, "expected '(' in cond clause")
		}

		if p.currentToken.TokenType == lexer.TokenTypeElse {
			break
		} else {
			//((predicate1) exps)
			test, err := p.parseExpression()
			if err != nil {
				return nil, NewParsingError(p.currentToken, err.Error())
			}

			exps := make([]Expression, 0)
			for p.currentToken.TokenType != lexer.TokenTypeRightParen {
				exp, err := p.parseExpression()
				if err != nil {
					return nil, NewParsingError(p.currentToken, err.Error())
				}
				exps = append(exps, exp)
			}
			var consequent Expression
			if len(exps) == 0 {
				return nil, NewParsingError(p.currentToken, "expected at least one expression in cond clause")
			} else if len(exps) == 1 {
				consequent = exps[0]
			} else {
				consequent = &BeginExpression{Expressions: exps, LeftParenToken: exps[0].Token()}
			}

			if !p.match(lexer.TokenTypeRightParen) {
				return nil, NewParsingError(p.currentToken, "expected ')' after cond clause")
			}

			if ifExp == nil {
				ifExp = &IfExpression{
					//LeftParenToken: test.Token(),
					Predicate:  test,
					Consequent: consequent,
				}
				currentIfExp = ifExp
			} else {
				newIfExp := &IfExpression{
					LeftParenToken: ifFirstToken,
					Predicate:      test,
					Consequent:     consequent,
				}
				currentIfExp.Alternative = newIfExp
				currentIfExp = newIfExp
			}
		}
	}

	if ifExp == nil {
		return nil, NewParsingError(p.currentToken, "expected at least one cond clause")
	}

	if p.currentToken.TokenType != lexer.TokenTypeElse {
		currentIfExp.Alternative = Void
	} else {
		firstToken := p.currentToken
		p.nextToken()

		exps := make([]Expression, 0)
		for p.currentToken.TokenType != lexer.TokenTypeRightParen {
			exp, err := p.parseExpression()
			if err != nil {
				return nil, NewParsingError(p.currentToken, err.Error())
			}
			exps = append(exps, exp)
		}

		var alternative Expression
		if len(exps) == 0 {
			return nil, NewParsingError(p.currentToken, "expected at least one expression in cond clause")
		} else if len(exps) == 1 {
			alternative = exps[0]
		} else {
			alternative = &BeginExpression{Expressions: exps, LeftParenToken: firstToken}
		}
		currentIfExp.Alternative = alternative

		if !p.match(lexer.TokenTypeRightParen) {
			return nil, NewParsingError(p.currentToken, "expected ')' at the end of else clause")
		}

	}

	if !p.match(lexer.TokenTypeRightParen) {
		return nil, NewParsingError(p.currentToken, "expected ')' after cond expression")
	}

	ifExp.LeftParenToken = ifFirstToken
	return ifExp, nil
}

func (p *Parser) parseSetExpression() (Expression, error) {
	p.nextToken()
	if p.currentToken.TokenType != lexer.TokenTypeIdentifier {
		return nil, NewParsingError(p.currentToken, "expected identifier after set!")
	}
	name := p.currentToken.Content

	p.nextToken()
	value, err := p.parseExpression()
	if err != nil {
		return nil, NewParsingError(p.currentToken, err.Error())
	}

	if !p.match(lexer.TokenTypeRightParen) {
		return nil, NewParsingError(p.currentToken, "expected ')' at the end of set expression")
	}
	return &SetExpression{
		Name:  name,
		Value: value,
	}, nil
}

func (p *Parser) parseBeginExpression() (Expression, error) {
	firstToken := p.currentToken
	p.nextToken()

	expressions := make([]Expression, 0)
	for p.currentToken.TokenType != lexer.TokenTypeRightParen {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, NewParsingError(p.currentToken, err.Error())
		}
		expressions = append(expressions, expr)
	}

	if len(expressions) == 0 {
		return nil, NewParsingError(p.currentToken, "expected at least one expression in begin")
	}

	p.nextToken()
	return &BeginExpression{
		LeftParenToken: firstToken,
		Expressions:    expressions,
	}, nil
}

func (p *Parser) parseGroupExpression() (Expression, error) {
	p.nextToken()

	switch p.currentToken.TokenType {
	case lexer.TokenTypeDefine:
		return p.parseDefineExpression()
	case lexer.TokenTypeLet:
		return p.parseLetExpression()
	case lexer.TokenTypeBegin:
		return p.parseBeginExpression()
	case lexer.TokenTypeSet:
		return p.parseSetExpression()
	case lexer.TokenTypeLambda:
		return p.parseLambdaExpression()
	case lexer.TokenTypeIf:
		return p.parseIfExpression()
	case lexer.TokenTypeCond:
		return p.parseCondExpression()
	case lexer.TokenTypeAnd:
		return p.parseCallExpression()
	case lexer.TokenTypeOr:
		return p.parseCallExpression()
	case lexer.TokenTypeNot:
		return p.parseCallExpression()
	case lexer.TokenTypeRightParen:
		p.nextToken()
		return &ListExpression{Elements: []Expression{}}, nil
	case lexer.TokenTypeDelay:
		return p.parseDelayExpression()
	case lexer.TokenTypeConsStream:
		return p.parseStreamExpression()
	default:
		// ( + 1 2 )
		// ( ( a b) )
		return p.parseCallExpression()
	}
}

func (p *Parser) parseStreamExpression() (Expression, error) {
	consStreamToken := p.currentToken
	p.nextToken()

	carExpression, err := p.parseExpression()
	if err != nil {
		return nil, NewParsingError(p.currentToken, err.Error())
	}

	cdrExpression, err := p.parseExpression()
	if err != nil {
		return nil, NewParsingError(p.currentToken, err.Error())
	}

	if !p.match(lexer.TokenTypeRightParen) {
		return nil, NewParsingError(p.currentToken, "expected ')' at the end of delay expression")
	}

	return &StreamExpression{
		ConsStreamToken: consStreamToken,
		CarExpression:   carExpression,
		CdrExpression:   cdrExpression,
	}, nil
}

func (p *Parser) parseDelayExpression() (Expression, error) {
	delayToken := p.currentToken
	p.nextToken()

	exp, err := p.parseExpression()
	if err != nil {
		return nil, NewParsingError(p.currentToken, err.Error())
	}
	if !p.match(lexer.TokenTypeRightParen) {
		return nil, NewParsingError(p.currentToken, "expected ')' at the end of delay expression")
	}
	return &DelayExpression{Expression: exp, DelayToken: delayToken}, nil
}

func (p *Parser) parsePrimitiveProcedure() (Expression, error) {
	exp := &PrimitiveProcedureExpression{Value: p.currentToken.Content, NameToken: p.currentToken}
	p.nextToken()
	return exp, nil
}

func (p *Parser) parseIdentifier() (Expression, error) {
	exp := &IdentifierExpression{Value: p.currentToken.Content, NameToken: p.currentToken}
	p.nextToken()
	return exp, nil
}

func (p *Parser) parseQuoteListExpression() (Expression, error) {
	// TODO: this implementation is not complete
	firstToken := p.currentToken

	//> '( '(a))
	//'('(a))

	//> (car '( '(a)) )
	//''(a)

	elements := make([]Expression, 0)
	for p.currentToken.TokenType != lexer.TokenTypeRightParen {
		switch p.currentToken.TokenType {
		case lexer.TokenTypeLeftParen:
			p.nextToken()
			element, err := p.parseQuoteListExpression()
			if err != nil {
				return nil, NewParsingError(p.currentToken, err.Error())
			}
			elements = append(elements, element)
		case lexer.TokenTypeNumber:
			element, err := p.parseNumber()
			if err != nil {
				return nil, NewParsingError(p.currentToken, err.Error())
			}
			elements = append(elements, element)
		case lexer.TokenTypeString:
			element, err := p.parseString()
			if err != nil {
				return nil, NewParsingError(p.currentToken, err.Error())
			}
			elements = append(elements, element)
		case lexer.TokenTypeEOF:
			return nil, NewParsingError(p.currentToken, fmt.Sprintf("unexpected token: %s", p.currentToken.TokenType))
		case lexer.TokenTypeInvalid:
			return nil, NewParsingError(p.currentToken, fmt.Sprintf("unexpected token: %s", p.currentToken.TokenType))
		default:
			element := &SymbolExpression{FirstToken: p.currentToken, Value: p.currentToken.Content}
			elements = append(elements, element)
			p.nextToken()
		}
	}

	p.nextToken()
	return &ListExpression{LeftParenToken: firstToken, Elements: elements}, nil
}

func (p *Parser) parseQuoteExpression() (Expression, error) {
	// the single quote can be used to denote lists or symbols.
	//	(define a 1)
	//(define b 2)
	//
	//(list a b)
	//(1 2)
	//
	//(list 'a 'b)
	//(a b)
	//
	//(list 'a b)
	//(a 2)
	//	(car '(a b c))
	//'a
	//
	//(cdr '(a b c))
	//'(b c)
	// '() -> null

	// TODO how to structure the quoted expression?
	// quoted expression can be a list or a symbol or a number or a string
	// for list, we can use CallExpression with Operator as nil
	// for symbol, we can use IdentifierExpression

	// 2025-09-28 we need to reconsider how to parse quote
	// i.e., ''a, is more like (cons ' 'a)
	quoteToken := p.currentToken
	p.nextToken()
	switch p.currentToken.TokenType {
	case lexer.TokenTypeLeftParen:
		p.nextToken()
		return p.parseQuoteListExpression()
	case lexer.TokenTypeNumber:
		return p.parseNumber()
	case lexer.TokenTypeString:
		return p.parseString()
	case lexer.TokenTypeEOF:
		return nil, NewParsingError(p.currentToken, fmt.Sprintf("unexpected token: %s", p.currentToken.TokenType))
	case lexer.TokenTypeRightParen:
		return nil, NewParsingError(p.currentToken, fmt.Sprintf("unexpected token: %s", p.currentToken.TokenType))
	case lexer.TokenTypeQuote:
		exp, err := p.parseQuoteExpression()
		if err != nil {
			return nil, NewParsingError(p.currentToken, err.Error())
		}

		return &NestedSymbolExpression{quoteToken, exp}, nil
	default:
		val := p.currentToken.Content
		p.nextToken()
		return &SymbolExpression{quoteToken, val}, nil
	}
}

func (p *Parser) parseExpression() (Expression, error) {
	switch p.currentToken.TokenType {
	case lexer.TokenTypeNumber:
		return p.parseNumber()
	case lexer.TokenTypeString:
		return p.parseString()
	case lexer.TokenTypeLeftParen:
		return p.parseGroupExpression()
	case lexer.TokenTypeEOF:
		return nil, NewParsingError(p.currentToken, "EOF")
	case lexer.TokenTypePlus:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeMinus:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeAsterisk:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeSlash:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeGreater:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeGreaterEqual:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeLess:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeLessEqual:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeAnd:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeOr:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeNot:
		return p.parsePrimitiveProcedure()
	case lexer.TokenTypeQuote:
		return p.parseQuoteExpression()
	case lexer.TokenTypeIdentifier:
		return p.parseIdentifier()
	case lexer.TokenTypeTrue:
		p.nextToken()
		return TrueLiteral, nil
	case lexer.TokenTypeFalse:
		p.nextToken()
		return FalseLiteral, nil
	case lexer.TokenTypeForce:
		return p.parsePrimitiveProcedure()

	default:
		return nil, NewParsingError(p.currentToken, fmt.Sprintf("unexpected token: %s", p.currentToken.TokenType))
	}
}
