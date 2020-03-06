package query

import (
	"github.com/x-foby/w3sql/ast"
	"github.com/x-foby/w3sql/source"
	"github.com/x-foby/w3sql/token"
)

// Query contains prepared AST-nodes
type Query struct {
	path      string
	fields    *ast.IdentList
	condition ast.Expr
	orderBy   *ast.OrderByStmtList
	limits    *ast.LimitsStmt
	source    *source.Source
}

// New returns new Query
func New(path string, fields *ast.IdentList, expr ast.Expr, orderBy *ast.OrderByStmtList, limits *ast.LimitsStmt) *Query {
	return &Query{path: path,
		fields:    fields,
		condition: expr,
		orderBy:   orderBy,
		limits:    limits,
	}
}

// WithSource set Source
func (q *Query) WithSource(s *source.Source) *Query {
	q.source = s
	return q
}

// Path returns path
func (q *Query) Path() string {
	return q.path
}

// Fields returns fields
func (q *Query) Fields() *ast.IdentList {
	return q.fields
}

// Condition return Query condition
func (q *Query) Condition() ast.Expr {
	return q.condition
}

// RewriteCondition set new condition for Quoery instead current condition
func (q *Query) RewriteCondition(cond ast.Expr) {
	q.condition = cond
}

// WrapCondition wrap current condition to new condition as y
func (q *Query) WrapCondition(x ast.Expr, operator token.Token) {
	if q.condition == nil {
		q.condition = x
	} else {
		q.condition = ast.NewBinaryExpr(operator, x, q.condition, 0)
	}
}
