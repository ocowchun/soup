package lexer

import (
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	input := `
(define if lambda let begin set! cond else and or square > < >= <= + - * / ' "hello" 123 45.67)
+abc -bcd *cd /de *123 /67 +123 -45.67 #t #f
. .a a.b
true false
`
	l := New(strings.NewReader(input))
	expectedTokens := []Token{
		{Content: "(", Line: 2, TokenType: TokenTypeLeftParen},
		{Content: "define", Line: 2, TokenType: TokenTypeDefine},
		{Content: "if", Line: 2, TokenType: TokenTypeIf},
		{Content: "lambda", Line: 2, TokenType: TokenTypeLambda},
		{Content: "let", Line: 2, TokenType: TokenTypeLet},
		{Content: "begin", Line: 2, TokenType: TokenTypeBegin},
		{Content: "set!", Line: 2, TokenType: TokenTypeSet},
		{Content: "cond", Line: 2, TokenType: TokenTypeCond},
		{Content: "else", Line: 2, TokenType: TokenTypeElse},
		{Content: "and", Line: 2, TokenType: TokenTypeAnd},
		{Content: "or", Line: 2, TokenType: TokenTypeOr},
		{Content: "square", Line: 2, TokenType: TokenTypeIdentifier},
		{Content: ">", Line: 2, TokenType: TokenTypeGreater},
		{Content: "<", Line: 2, TokenType: TokenTypeLess},
		{Content: ">=", Line: 2, TokenType: TokenTypeGreaterEqual},
		{Content: "<=", Line: 2, TokenType: TokenTypeLessEqual},
		{Content: "+", Line: 2, TokenType: TokenTypePlus},
		{Content: "-", Line: 2, TokenType: TokenTypeMinus},
		{Content: "*", Line: 2, TokenType: TokenTypeAsterisk},
		{Content: "/", Line: 2, TokenType: TokenTypeSlash},
		{Content: "'", Line: 2, TokenType: TokenTypeQuote},
		{Content: "hello", Line: 2, TokenType: TokenTypeString},
		{Content: "123", Line: 2, TokenType: TokenTypeNumber},
		{Content: "45.67", Line: 2, TokenType: TokenTypeNumber},
		{Content: ")", Line: 2, TokenType: TokenTypeRightParen},
		{Content: "+abc", Line: 3, TokenType: TokenTypeIdentifier},
		{Content: "-bcd", Line: 3, TokenType: TokenTypeIdentifier},
		{Content: "*cd", Line: 3, TokenType: TokenTypeIdentifier},
		{Content: "/de", Line: 3, TokenType: TokenTypeIdentifier},
		{Content: "*123", Line: 3, TokenType: TokenTypeIdentifier},
		{Content: "/67", Line: 3, TokenType: TokenTypeIdentifier},
		{Content: "+123", Line: 3, TokenType: TokenTypeNumber},
		{Content: "-45.67", Line: 3, TokenType: TokenTypeNumber},
		{Content: "#t", Line: 3, TokenType: TokenTypeTrue},
		{Content: "#f", Line: 3, TokenType: TokenTypeFalse},
		{Content: ".", Line: 4, TokenType: TokenTypeDot},
		{Content: ".a", Line: 4, TokenType: TokenTypeIdentifier},
		{Content: "a.b", Line: 4, TokenType: TokenTypeIdentifier},
		{Content: "true", Line: 5, TokenType: TokenTypeTrue},
		{Content: "false", Line: 5, TokenType: TokenTypeFalse},
		{Content: "", Line: 5, TokenType: TokenTypeEOF},
	}

	for i, expected := range expectedTokens {
		tok := l.NextToken()
		if tok.TokenType == TokenTypeInvalid {
			t.Fatalf("unexpected error at token %d: %v", i, tok.Content)
		}
		if tok != expected {
			t.Fatalf("unexpected token at %d: got %+v, want %+v", i, tok, expected)
		}
	}
}
