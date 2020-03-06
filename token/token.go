package token

import (
	"fmt"
)

// Token is type of some token
type Token int

// consts
const (
	ILLEGAL Token = iota
	EOF

	literalbeg
	IDENT  // x
	INT    // 123
	FLOAT  // 1.23
	STRING // "abc"
	LIST   // {1,2,3}
	PSEUDO // $count
	literalend

	operatorsbeg
	AND // &
	OR  // |

	EQL  // = (behave like IN for LIST)
	NEQ  // != (behave like NOT IN for LIST)
	LSS  // <
	LEQ  // <=
	GTR  // >
	GEQ  // >=
	LIKE // ~=
	operatorsend

	NOT // !

	LPAREN // (
	LBRACK // [
	LBRACE // {

	RPAREN // )
	RBRACK // ]
	RBRACE // }

	PLUS  // +
	MINUS // -

	COMMA // ,
	COLON // :
	AT    // @
	QUERY // ?
	QUO   // /
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",
	LIST:   "LIST",
	PSEUDO: "PSEUDO",

	AND: "&",
	OR:  "|",

	EQL:  "=",
	NEQ:  "!=",
	LSS:  "<",
	LEQ:  "<=",
	GTR:  ">",
	GEQ:  ">=",
	LIKE: "~=",

	NOT: "!",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",

	RPAREN: ")",
	RBRACK: "]",
	RBRACE: "}",

	PLUS:  "+",
	MINUS: "-",

	COMMA: ",",
	COLON: ":",
	AT:    "@",
	QUERY: "?",
	QUO:   "/",
}

// String returns the string corresponding to Token
func (t Token) String() string {
	if t < 0 || t >= Token(len(tokens)) {
		return fmt.Sprintf("undefined (%v)", int(t))
	}

	return tokens[t]
}

// Precedence returns the operator precedence of the binary operator op
func (t Token) Precedence() int {
	switch t {
	case OR:
		return 1
	case AND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ, LIKE:
		return 3

	default:
		return 0
	}
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
func (t Token) IsLiteral() bool { return literalbeg < t && t < literalend }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (t Token) IsOperator() bool { return operatorsbeg < t && t < operatorsend }
