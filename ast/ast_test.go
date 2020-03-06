package ast

import (
	"testing"

	"github.com/x-foby/w3sql/token"
)

func TestNewIdentList(t *testing.T) {
	foo := NewIdent("foo", 123)
	bar := NewIdent("bar", 321)
	identList := NewIdentList(foo)
	if identList == nil || len(*identList) != 1 || (*identList)[0] != foo || (*identList)[0].Pos() != (*identList)[0].pos || (*identList)[0].pos != 123 {
		t.Fail()
	}
	identList.Append(bar)
	if identList == nil || len(*identList) != 2 || (*identList)[1] != bar || (*identList)[1].Pos() != (*identList)[1].pos || (*identList)[1].pos != 321 {
		t.Fail()
	}
}

func TestNewOrderByStmtList(t *testing.T) {
	foo := NewIdent("foo", 123)
	fooDir := NewOrderByDir(OrderAsc, 123, token.PLUS)
	fooStmt := NewOrderByStmt(foo, fooDir)
	bar := NewIdent("bar", 321)
	barDir := NewOrderByDir(OrderDesc, 321, token.MINUS)
	barStmt := NewOrderByStmt(bar, barDir)

	list := NewOrderByStmtList(fooStmt)
	if list == nil || len(*list) != 1 || (*list)[0].Field != foo || (*list)[0].Direction != fooDir {
		t.Fail()
	}
	list.Append(barStmt)
	if list == nil || len(*list) != 2 || (*list)[1].Field != bar || (*list)[1].Direction != barDir {
		t.Fail()
	}
}

func TestNewLimitsStmt(t *testing.T) {
	from := NewConst("10", 123, token.INT)
	len := NewConst("", 321, token.EOF)
	l := NewLimitsStmt(from, len)
	if l.From == nil || l.From.Value != "10" || l.From.pos != l.From.Pos() || l.From.pos != 123 || l.From.tok != token.INT {
		t.Fail()
	}
	if l.Len == nil || l.Len.Value != "" || l.Len.pos != l.Len.Pos() || l.Len.pos != 321 || l.Len.tok != token.EOF {
		t.Fail()
	}
}

func TestNewExprList(t *testing.T) {
	foo := NewIdent("foo", 123)
	fooValue := NewConst("123", 321, token.STRING)
	fooExpr := NewBinaryExpr(token.EQL, foo, fooValue, 234)

	bar := NewIdent("bar", 567)
	barValue := NewConst("123", 765, token.STRING)
	barExpr := NewBinaryExpr(token.EQL, bar, barValue, 678)

	list := NewExprList(111, fooExpr)
	if list == nil || list.Pos() != 111 || len(list.Exprs) != 1 || list.Exprs[0].Pos() != fooExpr.Pos() {
		t.Fail()
	}
	list.Append(barExpr)
	if list == nil || list.Pos() != 111 || len(list.Exprs) != 2 || list.Exprs[1].Pos() != barExpr.Pos() {
		t.Fail()
	}
}

func TestIsFlat(t *testing.T) {
	foo := NewIdent("foo", 123)
	fooValue := NewConst("123", 321, token.STRING)
	fooExpr := NewBinaryExpr(token.EQL, foo, fooValue, 234)
	if !fooExpr.IsFlat() {
		t.Fail()
	}

	bar := NewIdent("bar", 567)
	barValue := NewConst("567", 765, token.STRING)
	barExpr := NewBinaryExpr(token.EQL, bar, barValue, 678)
	if !barExpr.IsFlat() {
		t.Fail()
	}

	fooAndbarExpr := NewBinaryExpr(token.AND, fooExpr, barExpr, 777)
	if !fooAndbarExpr.IsFlat() {
		t.Fail()
	}

	fooLeqbarExpr := NewBinaryExpr(token.LEQ, fooExpr, barExpr, 666)
	if fooLeqbarExpr.IsFlat() {
		t.Fail()
	}

	baz := NewIdent("baz", 567)
	bazExpr := NewBinaryExpr(token.EQL, baz, fooAndbarExpr, 678)
	if bazExpr.IsFlat() {
		t.Fail()
	}
}

func TestNewUnaryExpr(t *testing.T) {
	foo := NewIdent("foo", 123)
	expr := NewUnaryExpr(token.NOT, foo, 123)
	if expr == nil || expr.Pos() != 123 || expr.X != foo {
		t.Fail()
	}
}
