package w3sql

import "strings"

// Node is a AST-node
type Node interface {
	token() token
}

// Limits contains information about size results
type Limits struct {
	From int // offset
	Len  int // limit
}

// Query is AST
// Query constains:
//   Path - is identifiers set (example schema/table),
//   Fields - is like SQL FROM,
//   Condition - is like SQL WHERE,
//   Limits - is like SQL OFFSET and LIMIT
type Query struct {
	server    *Server
	path      *IdentList
	fields    *IdentList
	condition Expr
	limits    *Limits
}

// Limits return copy of Query limits
func (q *Query) Limits() *Limits {
	if q.limits == nil {
		return nil
	}

	limits := *q.limits
	return &limits
}

// Condition return Query condition
func (q *Query) Condition() Expr {
	return q.condition
}

// RewriteCondition set new condition for Quoery instead current condition
func (q *Query) RewriteCondition(cond Expr) {
	q.condition = cond
}

// WrapCondition wrap current condition to new condition as y
func (q *Query) WrapCondition(x Expr, operator token) {
	if q.condition == nil {
		q.condition = x
	} else {
		q.condition = BinaryExpr{X: x, Op: operator, Y: q.condition}
	}
}

// Path return full http-part
func (q Query) Path() string {
	if q.path == nil {
		return "/"
	}

	var path []string

	for _, ident := range *q.path {
		path = append(path, ident.Name)
	}

	return "/" + strings.Join(path, "/")
}

// SQL return full http-part
func (q Query) SQL(target string) (string, error) {
	return compile(q, target)
}

// Ident contains information about some identifier
type Ident struct {
	Pos  int
	Name string
	tok  token
}

// Token return token
func (i Ident) token() token { return i.tok }

// IdentList 1
type IdentList []*Ident

// Token return token
func (i IdentList) token() token { return LBRACE }

// Const contains information about some constant
type Const struct {
	Pos   int
	Value string
	tok   token
}

// Token return token
func (i Const) token() token { return i.tok }

// Expr is AST-node
type Expr interface {
	Node
}

// ExprList contains list of expressions
type ExprList struct {
	Pos   int
	Value []Expr
	tok   token
}

// Token return token
func (e ExprList) token() token { return e.tok }

// BinaryExpr contains X and Y as expressions and operator
type BinaryExpr struct {
	Pos int
	Op  token
	X   Expr
	Y   Expr
}

// Token return token
func (e BinaryExpr) token() token { return e.Op }

// UnaryExpr contains X as expressions and operator
type UnaryExpr struct {
	Pos int
	Op  token
	X   Expr
}

// Token return token
func (e UnaryExpr) token() token { return e.Op }
