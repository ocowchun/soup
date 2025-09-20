package lexer

import (
	"bufio"
	"fmt"
	"io"
)

type Lexer struct {
	scanner *bufio.Scanner
	line    string
	lineNo  int
	column  int
}

type TokenType uint8

const (
	TokenTypeNone TokenType = iota
	TokenTypeInvalid
	TokenTypeEOF
	TokenTypeNumber
	TokenTypeString
	TokenTypeIdentifier
	TokenTypeKeyword
	TokenTypeLeftParen
	TokenTypeRightParen
	TokenTypePlus
	TokenTypeMinus
	TokenTypeAsterisk
	TokenTypeSlash
	TokenTypeQuote
	TokenTypeDot
	TokenTypeIf
	TokenTypeDefine
	TokenTypeLambda
	TokenTypeLet
	TokenTypeBegin
	TokenTypeSet
	TokenTypeCond
	TokenTypeElse
	TokenTypeAnd
	TokenTypeOr
	TokenTypeTrue
	TokenTypeFalse
	TokenTypeLess
	TokenTypeGreater
	TokenTypeLessEqual
	TokenTypeGreaterEqual
)

func (t TokenType) String() string {
	switch t {
	case TokenTypeNone:
		return "None"
	case TokenTypeInvalid:
		return "Invalid"
	case TokenTypeEOF:
		return "EOF"
	case TokenTypeNumber:
		return "Number"
	case TokenTypeString:
		return "String"
	case TokenTypeIdentifier:
		return "Identifier"
	case TokenTypeKeyword:
		return "Keyword"
	case TokenTypeLeftParen:
		return "LeftParen"
	case TokenTypeRightParen:
		return "RightParen"
	case TokenTypePlus:
		return "Plus"
	case TokenTypeMinus:
		return "Minus"
	case TokenTypeAsterisk:
		return "Asterisk"
	case TokenTypeSlash:
		return "Slash"
	case TokenTypeQuote:
		return "Quote"
	case TokenTypeDot:
		return "Dot"
	case TokenTypeIf:
		return "If"
	case TokenTypeDefine:
		return "Define"
	case TokenTypeLambda:
		return "Lambda"
	case TokenTypeLet:
		return "Let"
	case TokenTypeBegin:
		return "Begin"
	case TokenTypeSet:
		return "Set!"
	case TokenTypeCond:
		return "Cond"
	case TokenTypeElse:
		return "Else"
	case TokenTypeAnd:
		return "And"
	case TokenTypeOr:
		return "Or"
	case TokenTypeTrue:
		return "True"
	case TokenTypeFalse:
		return "False"
	case TokenTypeLess:
		return "Less"
	case TokenTypeGreater:
		return "Greater"
	case TokenTypeLessEqual:
		return "LessEqual"
	case TokenTypeGreaterEqual:
		return "GreaterEqual"
	default:
		return "Unknown"
	}
}

type Token struct {
	Content   string
	Line      int
	TokenType TokenType
}

func New(reader io.Reader) *Lexer {
	scanner := bufio.NewScanner(reader)
	return &Lexer{
		scanner: scanner,
		line:    "",
		lineNo:  0,
		column:  0,
	}
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphabet(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

//func isAlpha(c byte) bool {
//	return c >= '0' && c <= '9'
//}

func (l *Lexer) readNumber(acceptDot bool) (string, error) {
	start := l.column - 1
	for l.column < len(l.line) && isDigit(l.line[l.column]) {
		l.column++
	}
	// TODO: handle float correctly
	if l.column < len(l.line) && l.line[l.column] == '.' {
		if !acceptDot {
			return "", fmt.Errorf("invalid character '.' in number at line %d, column %d", l.lineNo, l.column+1)
		}

		l.column++
		for l.column < len(l.line) && isDigit(l.line[l.column]) {
			l.column++
		}
	}

	if l.column < len(l.line) {
		firstChar := l.line[l.column]
		if firstChar != '(' && firstChar != ')' && !isSpaceOrNewline(firstChar) {
			return "", fmt.Errorf("invalid character '%c' after number at line %d, column %d", firstChar, l.lineNo, l.column+1)
		}
	}

	return l.line[start:l.column], nil
}

func (l *Lexer) skipWhitespace() {
	for l.column < len(l.line) && isSpaceOrNewline(l.line[l.column]) {
		l.column++
	}
}

func isSpaceOrNewline(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t'
}

// readNextLine reads the next line from the scanner.
// It returns false if there are no more lines to read.
func (l *Lexer) readNextLine() bool {
	if !l.scanner.Scan() {
		return false
	}
	l.line = l.scanner.Text()
	l.lineNo = l.lineNo + 1
	l.column = 0
	return true
}

var keywordMap = map[string]TokenType{
	"define": TokenTypeDefine,
	"if":     TokenTypeIf,
	"lambda": TokenTypeLambda,
	"let":    TokenTypeLet,
	"begin":  TokenTypeBegin,
	"set!":   TokenTypeSet,
	"cond":   TokenTypeCond,
	"else":   TokenTypeElse,
	"and":    TokenTypeAnd,
	"or":     TokenTypeOr,
	"true":   TokenTypeTrue,
	"false":  TokenTypeFalse,
}

func (l *Lexer) readIdentifierOrKeyword() (Token, error) {
	start := l.column - 1
	// can be identifier or keyword
	for l.column < len(l.line) && !isSpaceOrNewline(l.line[l.column]) && l.line[l.column] != '(' && l.line[l.column] != ')' {
		l.column++
	}

	content := l.line[start:l.column]

	if tokenType, ok := keywordMap[content]; ok {
		return Token{Content: content, Line: l.lineNo, TokenType: tokenType}, nil
	} else {
		return Token{Content: content, Line: l.lineNo, TokenType: TokenTypeIdentifier}, nil
	}
}

func (l *Lexer) readString() (Token, error) {
	//string can be multi-line
	start := l.column
	content := ""
	for l.column == len(l.line) || l.line[l.column] != '"' {
		if l.column == len(l.line) {
			content += l.line[start:l.column]

			// read next line
			if !l.readNextLine() {
				return Token{}, fmt.Errorf("unterminated string at line %d, column %d", l.lineNo, l.column)
			}
			// include newline in string
			start = 0
			l.column = 0

		} else {
			l.column++
		}
	}
	content += l.line[start:l.column]

	if l.column == len(l.line) || l.line[l.column] != '"' {
		return Token{}, fmt.Errorf("unterminated string at line %d, column %d", l.lineNo, l.column)
	}
	l.column++

	return Token{Content: content, Line: l.lineNo, TokenType: TokenTypeString}, nil
}

// skipComment skips the comment starting with `;` or `#` until the end of the line.
// It returns false if there are no more lines to read.
func (l *Lexer) skipComment() bool {
	// treat `#lang racket/base` as comment first
	//if l.column < len(l.line) && l.line[l.column] == '#' {
	//	return l.readNextLine()
	//}
	if l.column < len(l.line) && l.line[l.column] == ';' {
		return l.readNextLine()
	}

	return true
}

func isComment(c byte) bool {
	return c == ';'
}

func (l *Lexer) isLangDirective() bool {
	// #lang
	target := "#lang "
	for i := 0; i < len(target); i++ {
		if l.column+i >= len(l.line) || l.line[l.column+i] != target[i] {
			return false
		}
	}

	return true
}

func (l *Lexer) readSharp() (Token, error) {
	// TODO: handle other cases like #(123)
	start := l.column - 1
	for l.column < len(l.line) && !isSpaceOrNewline(l.line[l.column]) && l.line[l.column] != '(' && l.line[l.column] != ')' {
		l.column++
	}
	content := l.line[start:l.column]
	if content == "#t" || content == "#true" {
		return Token{Content: content, Line: l.lineNo, TokenType: TokenTypeTrue}, nil
	} else if content == "#f" || content == "#false" {
		return Token{Content: content, Line: l.lineNo, TokenType: TokenTypeFalse}, nil
	}

	return Token{}, fmt.Errorf("invalid token after #: %s at line %d, column %d", content, l.lineNo, start)
}

func (l *Lexer) NextToken() Token {
	for l.column == len(l.line) || isSpaceOrNewline(l.line[l.column]) || isComment(l.line[l.column]) || l.isLangDirective() {
		if l.column == len(l.line) || l.isLangDirective() {
			if !l.readNextLine() {
				return Token{TokenType: TokenTypeEOF, Line: l.lineNo}
			}
		}

		l.skipWhitespace()
		if !l.skipComment() {
			return Token{TokenType: TokenTypeEOF, Line: l.lineNo}
		}
	}

	content := ""
	firstChar := l.line[l.column]
	l.column++
	var nextChar byte
	hasNextChar := false
	if l.column < len(l.line) {
		nextChar = l.line[l.column]
		hasNextChar = true
	}
	tokenType := TokenTypeNone
	switch firstChar {
	case '(':
		content = "("
		tokenType = TokenTypeLeftParen
	case ')':
		content = ")"
		tokenType = TokenTypeRightParen
		//TODO: handle +foo, -bar, *baz, /qux, these are all valid identifiers in scheme
		// also *123, /123 are valid identifiers, but +123, -123 are not valid identifiers
	case '+':
		if hasNextChar && !isSpaceOrNewline(nextChar) {
			if isDigit(nextChar) {
				n, err := l.readNumber(true)
				if err != nil {
					return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
				}
				content = n
				tokenType = TokenTypeNumber
			} else {
				token, err := l.readIdentifierOrKeyword()
				if err != nil {
					return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
				}
				return token
			}
		} else {
			content = "+"
			tokenType = TokenTypePlus
		}
	case '-':
		if hasNextChar && !isSpaceOrNewline(nextChar) {
			if isDigit(nextChar) {
				n, err := l.readNumber(true)
				if err != nil {
					return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
				}
				content = n
				tokenType = TokenTypeNumber
			} else {
				token, err := l.readIdentifierOrKeyword()
				if err != nil {
					return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
				}
				return token
			}
		} else {
			content = "-"
			tokenType = TokenTypeMinus
		}
	case '*':
		if hasNextChar && !isSpaceOrNewline(nextChar) {
			token, err := l.readIdentifierOrKeyword()
			if err != nil {
				return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
			}
			return token
		} else {
			content = "*"
			tokenType = TokenTypeAsterisk
		}
	case '/':
		if hasNextChar && !isSpaceOrNewline(nextChar) {
			token, err := l.readIdentifierOrKeyword()
			if err != nil {
				return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
			}
			return token
		} else {
			content = "/"
			tokenType = TokenTypeSlash
		}
	case '"':
		token, err := l.readString()
		if err != nil {
			return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
		}
		return token
	case '\'':
		content = "'"
		tokenType = TokenTypeQuote
	case '<':
		content = "<"
		tokenType = TokenTypeLess
		if l.column < len(l.line) && l.line[l.column] == '=' {
			l.column++
			content = "<="
			tokenType = TokenTypeLessEqual
		}
	case '>':
		content = ">"
		tokenType = TokenTypeGreater
		if l.column < len(l.line) && l.line[l.column] == '=' {
			l.column++
			content = ">="
			tokenType = TokenTypeGreaterEqual
		}
	case '#':
		tok, err := l.readSharp()
		if err != nil {
			return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
		}
		return tok

	case '.':
		if hasNextChar && !isSpaceOrNewline(nextChar) {
			if isDigit(nextChar) {
				// .123
				n, err := l.readNumber(false)
				if err != nil {
					return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
				}
				content = fmt.Sprintf(".%s", n)
				tokenType = TokenTypeNumber
			} else if isAlphabet(nextChar) {
				token, err := l.readIdentifierOrKeyword()
				if err != nil {
					return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
				}
				content = token.Content
				tokenType = TokenTypeIdentifier
			} else {
				content = fmt.Sprintf("invalid character '%c' after . at line %d, column %d", nextChar, l.lineNo, l.column)
				return Token{Content: content, Line: l.lineNo, TokenType: TokenTypeInvalid}
			}
		} else {
			content = "."
			tokenType = TokenTypeDot
		}

	default:
		if isDigit(firstChar) {
			// do we need to handle 123a?
			n, err := l.readNumber(true)
			if err != nil {
				return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
			}
			content = n
			tokenType = TokenTypeNumber
		} else {
			token, err := l.readIdentifierOrKeyword()
			if err != nil {
				return Token{Content: err.Error(), Line: l.lineNo, TokenType: TokenTypeInvalid}
			}
			return token
		}
	}

	return Token{Content: content, Line: l.lineNo, TokenType: tokenType}
}
