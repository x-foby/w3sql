package ast

import "github.com/x-foby/w3sql/token"

// Node is AST-node
type Node interface {
	Pos() token.Pos
	Token() token.Token
}

// Ident contains information about some identifier
type Ident struct {
	Name string
	pos  token.Pos
}

// NewIdent returns new Ident
func NewIdent(name string, pos token.Pos) *Ident {
	return &Ident{Name: name, pos: pos}
}

// Pos return position
func (i *Ident) Pos() token.Pos { return i.pos }

// Token return token
func (i *Ident) Token() token.Token { return token.IDENT }

// IdentList contains *Ident's
type IdentList []*Ident

// NewIdentList returns new IdentList
func NewIdentList(idents ...*Ident) *IdentList {
	list := IdentList(idents)
	return &list
}

// Append append idents
func (i *IdentList) Append(idents ...*Ident) *IdentList {
	*i = append(*i, idents...)
	return i
}

// Const contains information about some constant
type Const struct {
	Value string
	tok   token.Token
	pos   token.Pos
}

// NewConst returns new Const
func NewConst(value string, pos token.Pos, tok token.Token) *Const {
	return &Const{Value: value, pos: pos, tok: tok}
}

// Pos return position
func (c *Const) Pos() token.Pos { return c.pos }

// Token return token
func (c *Const) Token() token.Token { return c.tok }

// OrderByDirType is a ASC or DESC
type OrderByDirType string

// consts
const (
	OrderAsc  OrderByDirType = "asc"
	OrderDesc OrderByDirType = "desc"
)

// OrderByDir contains information about order sort direction
type OrderByDir struct {
	Value OrderByDirType
	pos   token.Pos
	tok   token.Token
}

// NewOrderByDir returns new OrderByDir
func NewOrderByDir(value OrderByDirType, pos token.Pos, tok token.Token) *OrderByDir {
	return &OrderByDir{Value: value, pos: pos, tok: tok}
}

// OrderByStmt contains information about order field and direction
type OrderByStmt struct {
	Field     *Ident
	Direction *OrderByDir
}

// NewOrderByStmt returns new OrderByStmt
func NewOrderByStmt(field *Ident, direction *OrderByDir) *OrderByStmt {
	return &OrderByStmt{Field: field, Direction: direction}
}

// OrderByStmtList contains OrderByStmt's
type OrderByStmtList []*OrderByStmt

// NewOrderByStmtList returns new OrderByStmtList
func NewOrderByStmtList(stmts ...*OrderByStmt) *OrderByStmtList {
	list := OrderByStmtList(stmts)
	return &list
}

// Append append stmts
func (o *OrderByStmtList) Append(stmts ...*OrderByStmt) *OrderByStmtList {
	*o = append(*o, stmts...)
	return o
}

// LimitsStmt contains information about size results
type LimitsStmt struct {
	From *Const // offset
	Len  *Const // limit
}

// NewLimitsStmt returns new LimitsStmt
func NewLimitsStmt(from, len *Const) *LimitsStmt {
	return &LimitsStmt{From: from, Len: len}
}

// Expr is AST-node
type Expr interface {
	Node
}

// ExprList contains Expr's
type ExprList struct {
	Exprs []Expr
	pos   token.Pos
}

// NewExprList returns new ExprList
func NewExprList(pos token.Pos, exprs ...Expr) *ExprList {
	return &ExprList{Exprs: exprs, pos: pos}
}

// Append append exprs
func (e *ExprList) Append(exprs ...Expr) *ExprList {
	e.Exprs = append(e.Exprs, exprs...)
	return e
}

// Pos return position
func (e *ExprList) Pos() token.Pos { return e.pos }

// Token return token
func (e *ExprList) Token() token.Token { return token.LBRACE }

// BinaryExpr contains X and Y as expressions and operator
type BinaryExpr struct {
	X   Expr
	Y   Expr
	Op  token.Token
	pos token.Pos
}

// NewBinaryExpr returns new BinaryExpr
func NewBinaryExpr(operator token.Token, x, y Expr, pos token.Pos) *BinaryExpr {
	return &BinaryExpr{Op: operator, X: x, Y: y, pos: pos}
}

// Pos return position
func (e *BinaryExpr) Pos() token.Pos { return e.pos }

// Token return token
func (e *BinaryExpr) Token() token.Token { return e.Op }

// IsFlat return true if expression not contains nested BinaryExpr or all nested expressions
// is a BinaryExpr with AND operator
func (e *BinaryExpr) IsFlat() bool {
	x, xok := e.X.(*BinaryExpr)
	y, yok := e.Y.(*BinaryExpr)

	if !xok && !yok && e.Op == token.EQL {
		return true
	}
	if xok != yok {
		return false
	}
	if e.Op != token.AND {
		return false
	}
	return x.IsFlat() && y.IsFlat() && e.Op == token.AND
}

// UnaryExpr contains X as expressions and operator
type UnaryExpr struct {
	X   Expr
	Op  token.Token
	pos token.Pos
}

// NewUnaryExpr returns new UnaryExpr
func NewUnaryExpr(operator token.Token, x Expr, pos token.Pos) *UnaryExpr {
	return &UnaryExpr{Op: operator, X: x, pos: pos}
}

// Pos return position
func (e *UnaryExpr) Pos() token.Pos { return e.pos }

// Token return token
func (e *UnaryExpr) Token() token.Token { return e.Op }
