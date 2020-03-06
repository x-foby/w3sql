package scanner

import (
	"testing"

	"github.com/x-foby/w3sql/token"
)

var cases = []struct {
	Name string
	Src  string
	Pos  token.Pos
	Tok  token.Token
	Lit  string
}{
	{Name: "EOF", Src: "", Pos: 0, Tok: token.EOF, Lit: ""},

	{Name: "Identificator", Src: "foo", Pos: 0, Tok: token.IDENT, Lit: "foo"},
	{Name: "Identificator", Src: "foo.bar", Pos: 0, Tok: token.IDENT, Lit: "foo.bar"},
	{Name: "Identificator", Src: "foo1", Pos: 0, Tok: token.IDENT, Lit: "foo1"},
	{Name: "Identificator", Src: "foo_1", Pos: 0, Tok: token.IDENT, Lit: "foo_1"},
	{Name: "Identificator", Src: "foo_1.bar2_1", Pos: 0, Tok: token.IDENT, Lit: "foo_1.bar2_1"},
	{Name: "Identificator", Src: "foo-1.bar2_1", Pos: 0, Tok: token.IDENT, Lit: "foo"},
	{Name: "Identificator", Src: "  foo", Pos: 2, Tok: token.IDENT, Lit: "foo"},

	{Name: "Integer", Src: "123", Pos: 0, Tok: token.INT, Lit: "123"},
	{Name: "Integer", Src: "123foo", Pos: 0, Tok: token.INT, Lit: "123"},
	{Name: "Integer", Src: "123+", Pos: 0, Tok: token.INT, Lit: "123"},
	{Name: "Integer", Src: " 123+", Pos: 1, Tok: token.INT, Lit: "123"},

	{Name: "Float", Src: "1.23", Pos: 0, Tok: token.FLOAT, Lit: "1.23"},
	{Name: "Float", Src: "1.23foo", Pos: 0, Tok: token.FLOAT, Lit: "1.23"},
	{Name: "Float", Src: "1.foo23", Pos: 0, Tok: token.FLOAT, Lit: "1."},
	{Name: "Float", Src: ".23", Pos: 0, Tok: token.FLOAT, Lit: ".23"},
	{Name: "Float", Src: "1.23+", Pos: 0, Tok: token.FLOAT, Lit: "1.23"},
	{Name: "Float", Src: " 1.23+", Pos: 1, Tok: token.FLOAT, Lit: "1.23"},
	{Name: "Float", Src: "1.23.45", Pos: 0, Tok: token.FLOAT, Lit: "1.23"},

	{Name: "String", Src: `"foo"`, Pos: 0, Tok: token.STRING, Lit: "foo"},
	{Name: "String", Src: ` "foo"`, Pos: 1, Tok: token.STRING, Lit: "foo"},
	{Name: "String", Src: `"foo`, Pos: 4, Tok: token.ILLEGAL, Lit: ""},

	{Name: "Pseudo field", Src: "$foo", Pos: 0, Tok: token.PSEUDO, Lit: "foo"},
	{Name: "Pseudo field", Src: " $foo", Pos: 1, Tok: token.PSEUDO, Lit: "foo"},
	{Name: "Pseudo field", Src: "$foo ", Pos: 0, Tok: token.PSEUDO, Lit: "foo"},
	{Name: "Pseudo field", Src: "$", Pos: 1, Tok: token.ILLEGAL, Lit: ""},
	{Name: "Pseudo field", Src: " $", Pos: 2, Tok: token.ILLEGAL, Lit: ""},
	{Name: "Pseudo field", Src: " $ ", Pos: 2, Tok: token.ILLEGAL, Lit: ""},

	{Name: "Colon", Src: ":", Pos: 0, Tok: token.COLON, Lit: ""},
	{Name: "Comma", Src: ",", Pos: 0, Tok: token.COMMA, Lit: ""},
	{Name: "Left parenthesis", Src: "(", Pos: 0, Tok: token.LPAREN, Lit: ""},
	{Name: "Right parenthesis", Src: ")", Pos: 0, Tok: token.RPAREN, Lit: ""},
	{Name: "Left brace", Src: "{", Pos: 0, Tok: token.LBRACE, Lit: ""},
	{Name: "Right brace", Src: "}", Pos: 0, Tok: token.RBRACE, Lit: ""},
	{Name: "Left square brack", Src: "[", Pos: 0, Tok: token.LBRACK, Lit: ""},
	{Name: "Right square brack", Src: "]", Pos: 0, Tok: token.RBRACK, Lit: ""},

	{Name: "And", Src: "&", Pos: 0, Tok: token.AND, Lit: ""},
	{Name: "Or", Src: "|", Pos: 0, Tok: token.OR, Lit: ""},
	{Name: "Not", Src: "!", Pos: 0, Tok: token.NOT, Lit: ""},

	{Name: "At", Src: "@", Pos: 0, Tok: token.AT, Lit: ""},
	{Name: "Query", Src: "?", Pos: 0, Tok: token.QUERY, Lit: ""},
	{Name: "Quo", Src: "/", Pos: 0, Tok: token.QUO, Lit: ""},
	{Name: "Sort ascending", Src: "+", Pos: 0, Tok: token.PLUS, Lit: ""},
	{Name: "Sort descending", Src: "-", Pos: 0, Tok: token.MINUS, Lit: ""},
	{Name: "Equal", Src: "=", Pos: 0, Tok: token.EQL, Lit: ""},
	{Name: "Not equal", Src: "!=", Pos: 0, Tok: token.NEQ, Lit: ""},
	{Name: "Less", Src: "<", Pos: 0, Tok: token.LSS, Lit: ""},
	{Name: "Less or equal", Src: "<=", Pos: 0, Tok: token.LEQ, Lit: ""},
	{Name: "Greater", Src: ">", Pos: 0, Tok: token.GTR, Lit: ""},
	{Name: "Greater or equal", Src: ">=", Pos: 0, Tok: token.GEQ, Lit: ""},
}

func TestScan(t *testing.T) {
	var s Scanner
	for _, c := range cases {
		s.Init([]rune(c.Src))
		t.Run(c.Name, func(t *testing.T) {
			pos, tok, lit := s.Scan()
			if pos != c.Pos {
				t.Errorf("expected pos: %v, got: %v", c.Pos, pos)
				t.Fail()
			}
			if tok != c.Tok {
				t.Errorf("expected token: %q, got: %q", c.Tok, tok)
				t.Fail()
			}
			if lit != c.Lit {
				t.Errorf("expected literal: %q, got: %q", c.Lit, lit)
				t.Fail()
			}
		})
	}
}

func TestPeek(t *testing.T) {
	var s Scanner
	if ch := s.peek(); ch != -1 {
		t.Errorf("expected: %q, got: %q", "EOF", ch)
		t.Fail()
	}
	s.Init([]rune(""))
	if ch := s.peek(); ch != -1 {
		t.Errorf("expected: %q, got: %q", "EOF", ch)
		t.Fail()
	}
	s.Init([]rune("f"))
	if ch := s.peek(); ch != -1 {
		t.Errorf("expected: %q, got: %q", "EOF", ch)
		t.Fail()
	}
	s.Init([]rune("foo"))
	if ch := s.peek(); ch != 'o' {
		t.Errorf("expected: %q, got: %q", "o", ch)
		t.Fail()
	}
}
