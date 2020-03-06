package parser

import (
	"reflect"
	"testing"

	"github.com/x-foby/w3sql/ast"
	"github.com/x-foby/w3sql/token"
)

var cases = []struct {
	Name    string
	Src     string
	Path    string
	Fields  *ast.IdentList
	Expr    ast.Expr
	OrderBy *ast.OrderByStmtList
	Limit   *ast.LimitsStmt
}{
	{Name: "Short path", Src: "/foo", Path: "foo"},
	{Name: "Long path without /", Src: "/foo/bar/baz", Path: "foo/bar/baz"},
	{Name: "Long path with /", Src: "/foo/bar/baz/", Path: "foo/bar/baz"},

	{Name: "Short path with field", Src: "/@foo", Path: "foo"},
	{Name: "Short path with field", Src: "/field@foo", Path: "foo", Fields: ast.NewIdentList(ast.NewIdent("field", 1))},
	{Name: "Short path with 2 fields", Src: "/field1,field2@foo", Path: "foo", Fields: ast.NewIdentList(ast.NewIdent("field1", 1), ast.NewIdent("field2", 8))},
	{Name: "Short path with 3 fields", Src: "/field1,field2,field3@foo", Path: "foo", Fields: ast.NewIdentList(ast.NewIdent("field1", 1), ast.NewIdent("field2", 8), ast.NewIdent("field3", 15))},

	{Name: "Query. 1 expr", Src: `/foo?a="b"`, Path: "foo", Expr: ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6)},
	{Name: "Query. 1 expr (negative int)", Src: `/foo?a=-1`, Path: "foo", Expr: ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewUnaryExpr(token.MINUS, ast.NewConst("1", 8, token.INT), 7), 6)},
	{
		Name: "Query. 3 expr",
		Src:  `/foo?a="b"&b="a"`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.AND,
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6),
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 11), ast.NewConst("a", 13, token.STRING), 12),
			10,
		),
	},
	{
		Name: "Query. Sequential precedences",
		Src:  `/foo?a="b"&b="a"|b="b"`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.OR,
			ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 11), ast.NewConst("a", 13, token.STRING), 12),
				10,
			),
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 17), ast.NewConst("b", 19, token.STRING), 18),
			16,
		),
	},
	{
		Name: "Query. Reversed precedences",
		Src:  `/foo?a="b"|b="a"&a="a"`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.OR,
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6),
			ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 11), ast.NewConst("a", 13, token.STRING), 12),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 17), ast.NewConst("a", 19, token.STRING), 18),
				16,
			),
			10,
		),
	},
	{
		Name: "Query. Paren first",
		Src:  `/foo?(a="b")&b="a"`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.AND,
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 6), ast.NewConst("b", 8, token.STRING), 7),
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 13), ast.NewConst("a", 15, token.STRING), 14),
			12,
		),
	},
	{
		Name: "Query. Paren last",
		Src:  `/foo?a="b"&(b="a")`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.AND,
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6),
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
			10,
		),
	},
	{
		Name: "Query. Paren",
		Src:  `/foo?a=(b&"b")`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.EQL,
			ast.NewIdent("a", 5),
			ast.NewBinaryExpr(token.AND, ast.NewIdent("b", 8), ast.NewConst("b", 10, token.STRING), 9),
			6,
		),
	},
	{
		Name: "Query. List",
		Src:  `/foo?a={1,2,3}`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.EQL,
			ast.NewIdent("a", 5),
			ast.NewExprList(
				7,
				ast.NewConst("1", 8, token.INT),
				ast.NewConst("2", 10, token.INT),
				ast.NewConst("3", 12, token.INT),
			),
			6,
		),
	},
	{
		Name: "Query. List",
		Src:  `/foo?a={1,2,3}|b=true`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.OR,
			ast.NewBinaryExpr(
				token.EQL,
				ast.NewIdent("a", 5),
				ast.NewExprList(
					7,
					ast.NewConst("1", 8, token.INT),
					ast.NewConst("2", 10, token.INT),
					ast.NewConst("3", 12, token.INT),
				),
				6,
			),
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 15), ast.NewIdent("true", 17), 16),
			14,
		),
	},
	{
		Name: "Query. JSON",
		Src:  `/foo?a={a="b"}`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.EQL,
			ast.NewIdent("a", 5),
			ast.NewExprList(
				7,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 8), ast.NewConst("b", 10, token.STRING), 9),
			),
			6,
		),
	},
	{
		Name: "Query. JSON",
		Src:  `/foo?a={a="b",b="a"}`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.EQL,
			ast.NewIdent("a", 5),
			ast.NewExprList(
				7,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 8), ast.NewConst("b", 10, token.STRING), 9),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 14), ast.NewConst("a", 16, token.STRING), 15),
			),
			6,
		),
	},
	{
		Name: "Query. JSON or",
		Src:  `/foo?a={a="b",b>="a"}|b=true`,
		Path: "foo",
		Expr: ast.NewBinaryExpr(
			token.OR,
			ast.NewBinaryExpr(
				token.EQL,
				ast.NewIdent("a", 5),
				ast.NewExprList(
					7,
					ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 8), ast.NewConst("b", 10, token.STRING), 9),
					ast.NewBinaryExpr(token.GEQ, ast.NewIdent("b", 14), ast.NewConst("a", 17, token.STRING), 15),
				),
				6,
			),
			ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 22), ast.NewIdent("true", 24), 23),
			21,
		),
	},
	{
		Name: "Sort. One field",
		Src:  "/foo:+a",
		Path: "foo",
		OrderBy: ast.NewOrderByStmtList(
			ast.NewOrderByStmt(
				ast.NewIdent("a", 6),
				ast.NewOrderByDir(ast.OrderAsc, 5, token.PLUS),
			),
		),
	},
	{
		Name: "Sort. Two field",
		Src:  "/foo:+a,-b",
		Path: "foo",
		OrderBy: ast.NewOrderByStmtList(
			ast.NewOrderByStmt(
				ast.NewIdent("a", 6),
				ast.NewOrderByDir(ast.OrderAsc, 5, token.PLUS),
			),
			ast.NewOrderByStmt(
				ast.NewIdent("6", 9),
				ast.NewOrderByDir(ast.OrderDesc, 8, token.MINUS),
			),
		),
	},
	{
		Name:  "Limits. No",
		Src:   "/foo[:]",
		Path:  "foo",
		Limit: ast.NewLimitsStmt(nil, nil),
	},
	{
		Name:  "Limits. No limit",
		Src:   "/foo[1:]",
		Path:  "foo",
		Limit: ast.NewLimitsStmt(ast.NewConst("1", 5, token.INT), nil),
	},
	{
		Name:  "Limits. No Offset",
		Src:   "/foo[:1]",
		Path:  "foo",
		Limit: ast.NewLimitsStmt(nil, ast.NewConst("1", 6, token.INT)),
	},
	{
		Name:  "Limits",
		Src:   "/foo[2:1]",
		Path:  "foo",
		Limit: ast.NewLimitsStmt(ast.NewConst("2", 5, token.INT), ast.NewConst("1", 7, token.INT)),
	},
	{
		Name: "Sort and limit",
		Src:  "/foo:+a[2:1]",
		Path: "foo",
		OrderBy: ast.NewOrderByStmtList(
			ast.NewOrderByStmt(
				ast.NewIdent("a", 6),
				ast.NewOrderByDir(ast.OrderAsc, 5, token.PLUS),
			),
		),
		Limit: ast.NewLimitsStmt(ast.NewConst("2", 8, token.INT), ast.NewConst("1", 10, token.INT)),
	},
}

func TestParse(t *testing.T) {
	p := New()

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			query, err := p.Parse(c.Src)
			if err != nil {
				t.Errorf("expected err: %v, got: %v", nil, err)
				t.FailNow()
			}

			if path := query.Path(); path != c.Path {
				t.Errorf("expected path: %v, got: %v", c.Path, path)
				t.Fail()
			}

			if fields := query.Fields(); !reflect.DeepEqual(c.Fields, fields) {
				t.Errorf("expected fields: %v, got: %v", c.Fields, fields)
				t.Fail()
			}

			if expr := query.Condition(); !reflect.DeepEqual(c.Expr, expr) {
				t.Errorf("expected expr: %v, got: %v", c.Expr, expr)
				t.Fail()
			}
		})
	}
}
