package w3sql

import (
	"fmt"
)

// Token is type of some token
type token int

// consts
const (
	ILLEGAL token = iota
	EOF

	literalbeg
	IDENT  // x
	INT    // 123
	FLOAT  // 1.23
	STRING // "abc"
	LIST   // {1,2,3}
	literalend

	operatorsbeg
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	AND // &
	OR  // |

	EQL // = (behave like IN for LIST)
	NEQ // != (behave like NOT IN for LIST)
	LSS // <
	LEQ // <=
	GTR // >
	GEQ // >=
	NOT // !
	// TODO like
	// LIKE // ~=
	operatorsend

	LPAREN // (
	LBRACK // [
	LBRACE // {

	RPAREN // )
	RBRACK // ]
	RBRACE // }

	COMMA // ,
	COLON // :
	AT    // @
	QUERY // ?
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",
	LIST:   "LIST",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	AND: "&",
	OR:  "|",

	EQL: "=",
	NEQ: "!=",
	LSS: "<",
	LEQ: "<=",
	GTR: ">",
	GEQ: ">=",
	NOT: "!",
	// LIKE: "~=",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",

	RPAREN: ")",
	RBRACK: "]",
	RBRACE: "}",

	COMMA: ",",
	COLON: ":",
	AT:    "@",
	QUERY: "?",
}

// String returns the string corresponding to Token
func (tok token) String() string {
	if tok < 0 || tok >= token(len(tokens)) {
		return fmt.Sprintf("undefined (%v)", int(tok))
	}

	return tokens[tok]
}

// Precedence returns the operator precedence of the binary operator op
func (tok token) Precedence() int {
	switch tok {
	case OR:
		return 1
	case AND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ /*, LIKE*/ :
		return 3
	case ADD, SUB:
		return 4
	case MUL, QUO, REM:
		return 5

	default:
		return 0
	}
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
func (tok token) IsLiteral() bool { return literalbeg < tok && tok < literalend }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
func (tok token) IsOperator() bool { return operatorsbeg < tok && tok < operatorsend }
