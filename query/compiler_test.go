package query

import (
	"testing"

	"github.com/x-foby/w3sql/ast"
	"github.com/x-foby/w3sql/source"
	"github.com/x-foby/w3sql/token"
)

var cases = []struct {
	Name   string
	Target string
	Query  *Query
	Result string
}{
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("colA", 5), ast.NewConst("b", 7, token.STRING), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewIdent("true", 14), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "colA", "col_a", false),
					source.NewCol(source.TypeBool, "b", "b", false),
				),
			},
		},
		Result: "select * from table where col_a = 'b' and b = true",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.NEQ, ast.NewIdent("colA", 5), ast.NewConst("b", 7, token.STRING), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewIdent("true", 14), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "colA", "col_a", false),
					source.NewCol(source.TypeBool, "b", "b", false),
				),
			},
		},
		Result: "select * from table where col_a != 'b' and b = true",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.NEQ, ast.NewIdent("colA", 5), ast.NewExprList(6, ast.NewConst("b", 7, token.STRING), ast.NewConst("a", 7, token.STRING)), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewIdent("true", 14), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "colA", "col_a", false),
					source.NewCol(source.TypeBool, "b", "b", false),
				),
			},
		},
		Result: "select * from table where col_a not in ('b', 'a') and b = true",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
				10,
			),
			orderBy: ast.NewOrderByStmtList(
				ast.NewOrderByStmt(ast.NewIdent("a", 5), ast.NewOrderByDir(ast.OrderAsc, 4, token.PLUS)),
				ast.NewOrderByStmt(ast.NewIdent("b", 8), ast.NewOrderByDir(ast.OrderDesc, 7, token.MINUS)),
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "a", "a", false),
					source.NewCol(source.TypeString, "b", "b", false),
				),
			},
		},
		Result: "select * from table where a = 'b' and b = 'a' order by a asc, b desc",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			orderBy: ast.NewOrderByStmtList(
				ast.NewOrderByStmt(ast.NewIdent("a.b", 5), ast.NewOrderByDir(ast.OrderAsc, 4, token.PLUS)),
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeObject, "a", "a", false).WithChildren(source.NewCols(
						source.NewCol(source.TypeNumber, "b", "b", false),
					)),
					source.NewCol(source.TypeString, "b", "b", false),
				),
			},
		},
		Result: "select * from table order by (a #>> '{b}')::numeric asc",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.LSS, ast.NewIdent("a", 5), ast.NewConst("1", 6, token.INT), 7),
				ast.NewBinaryExpr(
					token.AND,
					ast.NewBinaryExpr(token.LEQ, ast.NewIdent("b", 8), ast.NewConst("2", 9, token.INT), 10),
					ast.NewBinaryExpr(
						token.AND,
						ast.NewBinaryExpr(token.GTR, ast.NewIdent("c", 11), ast.NewConst("3", 12, token.INT), 13),
						ast.NewBinaryExpr(token.GEQ, ast.NewIdent("d", 14), ast.NewConst("4", 15, token.INT), 16),
						17,
					),
					18,
				),
				19,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeNumber, "a", "a", false),
					source.NewCol(source.TypeNumber, "b", "b", false),
					source.NewCol(source.TypeNumber, "c", "c", false),
					source.NewCol(source.TypeNumber, "d", "d", false),
				),
			},
		},
		Result: "select * from table where a < 1 and b <= 2 and c > 3 and d >= 4",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			fields: ast.NewIdentList(ast.NewIdent("a", 1)),
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "a", "a", false),
					source.NewCol(source.TypeString, "b", "b", false),
				),
			},
		},
		Result: "select a from table where a = 'b' and b = 'a'",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			fields: ast.NewIdentList(ast.NewIdent("a", 1), ast.NewIdent("b", 3)),
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewConst("b", 7, token.STRING), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "a", "a", false),
					source.NewCol(source.TypeString, "b", "b", false),
				),
			},
		},
		Result: "select a, b from table where a = 'b' and b = 'a'",
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			fields: ast.NewIdentList(ast.NewIdent("a", 1), ast.NewIdent("b", 3)),
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 5), ast.NewExprList(6, ast.NewConst("a", 14, token.STRING), ast.NewConst("b", 16, token.STRING)), 6),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "a", "a", false),
					source.NewCol(source.TypeString, "b", "b", true),
				),
			},
		},
		Result: `select a, b from table where a in ('a', 'b') and b @> '"a"'`,
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			fields: ast.NewIdentList(ast.NewIdent("a", 1), ast.NewIdent("b", 3)),
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(
					token.EQL,
					ast.NewIdent("a", 5),
					ast.NewExprList(6, ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 5), ast.NewConst("b", 16, token.STRING), 6)),
					16,
				),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeObject, "a", "a", true).WithChildren(source.NewCols(
						source.NewCol(source.TypeString, "b", "b", false),
					)),
					source.NewCol(source.TypeString, "b", "b", true),
				),
			},
		},
		Result: `select a, b from table where exists (select 1 from (select jsonb_array_elements(a::jsonb) item) q where (q.item #>> '{b}')::text = 'b'::text) and b @> '"a"'`,
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			fields: ast.NewIdentList(ast.NewIdent("a", 1), ast.NewIdent("b", 3)),
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(
					token.EQL,
					ast.NewIdent("a", 5),
					ast.NewExprList(6, ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 5), ast.NewConst("b", 16, token.STRING), 6)),
					16,
				),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeObject, "a", "a", false).WithChildren(source.NewCols(
						source.NewCol(source.TypeString, "b", "b", false),
					)),
					source.NewCol(source.TypeString, "b", "b", true),
				),
			},
		},
		Result: `select a, b from table where (a #>> '{b}')::text = 'b'::text and b @> '"a"'`,
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			fields: ast.NewIdentList(ast.NewIdent("a", 1), ast.NewIdent("b", 3)),
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(
					token.EQL,
					ast.NewIdent("a", 5),
					ast.NewExprList(6,
						ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 5), ast.NewConst("b", 16, token.STRING), 6),
						ast.NewBinaryExpr(token.EQL, ast.NewIdent("c.d", 5), ast.NewConst("4", 16, token.FLOAT), 6),
					),
					16,
				),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 12), ast.NewConst("a", 14, token.STRING), 13),
				10,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "b", "b", true),
					source.NewCol(source.TypeObject, "a", "a", true).WithChildren(source.NewCols(
						source.NewCol(source.TypeString, "b", "b", false),
						source.NewCol(source.TypeObject, "c", "c", true).WithChildren(source.NewCols(
							source.NewCol(source.TypeNumber, "d", "d", false),
						)),
					)),
				),
			},
		},
		Result: `select a, b from table where exists (select 1 from (select jsonb_array_elements(a::jsonb) item) q where (q.item #>> '{b}')::text = 'b'::text and (q.item #>> '{c,d}')::numeric = 4::numeric) and b @> '"a"'`,
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(
					token.OR,
					ast.NewBinaryExpr(
						token.EQL,
						ast.NewIdent("a", 6),
						ast.NewExprList(8, ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 9), ast.NewConst("b", 11, token.STRING), 10)),
						7,
					),
					ast.NewBinaryExpr(
						token.EQL,
						ast.NewIdent("a", 16),
						ast.NewExprList(18, ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 19), ast.NewConst("c", 21, token.STRING), 20)),
						17,
					),
					15,
				),
				ast.NewBinaryExpr(
					token.AND,
					ast.NewBinaryExpr(
						token.EQL,
						ast.NewIdent("a", 27),
						ast.NewExprList(29, ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 30), ast.NewConst("1", 32, token.INT), 31)),
						28,
					),
					ast.NewBinaryExpr(
						token.EQL,
						ast.NewIdent("b", 35),
						ast.NewIdent("true", 37),
						36,
					),
					34,
				),
				26,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeBool, "b", "b", false),
					source.NewCol(source.TypeObject, "a", "a", true).WithChildren(source.NewCols(
						source.NewCol(source.TypeString, "a", "a", false),
						source.NewCol(source.TypeNumber, "b", "b", false),
					)),
				),
			},
		},
		Result: `select * from table where (exists (select 1 from (select jsonb_array_elements(a::jsonb) item) q where (q.item #>> '{a}')::text = 'b'::text) or exists (select 1 from (select jsonb_array_elements(a::jsonb) item) q where (q.item #>> '{a}')::text = 'c'::text)) and exists (select 1 from (select jsonb_array_elements(a::jsonb) item) q where (q.item #>> '{b}')::numeric = 1::numeric) and b = true`,
	},
	{
		Name:   "Simple",
		Target: "table",
		Query: &Query{
			condition: ast.NewBinaryExpr(
				token.AND,
				ast.NewBinaryExpr(
					token.OR,
					ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 6), ast.NewConst("a", 8, token.STRING), 7),
					ast.NewBinaryExpr(
						token.OR,
						ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 12), ast.NewConst("b", 14, token.STRING), 13),
						ast.NewBinaryExpr(token.EQL, ast.NewIdent("a", 18), ast.NewConst("c", 20, token.STRING), 19),
						17,
					),
					11,
				),
				ast.NewBinaryExpr(token.EQL, ast.NewIdent("b", 25), ast.NewConst("a", 27, token.STRING), 26),
				24,
			),
			source: &source.Source{
				Cols: source.NewCols(
					source.NewCol(source.TypeString, "a", "a", false),
					source.NewCol(source.TypeString, "b", "b", false),
				),
			},
		},
		Result: `select * from table where (a = 'a' or a = 'b' or a = 'c') and b = 'a'`,
	},
}

func TestCompile(t *testing.T) {
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			sql, err := c.Query.Compile(c.Target)
			if err != nil {
				t.Errorf("expected err: %v, got: %v", nil, err)
				t.FailNow()
			}

			if sql != c.Result {
				t.Errorf("expected: %v, got: %v", c.Result, sql)
				t.Fail()
			}
		})
	}
}
