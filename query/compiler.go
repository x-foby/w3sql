package query

import (
	"errors"
	"fmt"
	"strings"

	"github.com/x-foby/w3sql/ast"
	"github.com/x-foby/w3sql/source"
	"github.com/x-foby/w3sql/token"
)

// Compile returns sql-query
func (q *Query) Compile(target string) (string, error) {
	if q.source == nil {
		return "", errors.New("source is not defined")
	}

	var (
		parts                                          []string
		selectStmt, whereStmt, orderByStmt, limitsStmt string
		err                                            error
	)
	selectStmt, err = q.compileSelect()
	if err != nil {
		return "", err
	}
	if selectStmt != "" {
		parts = append(parts, "select", selectStmt)
	}
	parts = append(parts, "from", target)
	whereStmt, err = q.compileWhere()
	if err != nil {
		return "", err
	}
	if whereStmt != "" {
		parts = append(parts, "where", whereStmt)
	}
	orderByStmt, err = q.compileOrderBy()
	if err != nil {
		return "", err
	}
	if orderByStmt != "" {
		parts = append(parts, "order by", orderByStmt)
	}
	limitsStmt, err = q.compileLimits()
	if err != nil {
		return "", err
	}
	if limitsStmt != "" {
		parts = append(parts, limitsStmt)
	}
	return strings.Join(parts, " "), nil
}

func (q *Query) compileSelect() (string, error) {
	if q.fields == nil || len(*q.fields) == 0 {
		return "*", nil
	}
	fields := make([]string, len(*q.fields))
	for i, f := range *q.fields {
		column := q.source.Cols.ByName(f.Name)
		if column == nil {
			return "", q.notDefined(f.Name, f.Pos())
		}
		fields[i] = column.DBName
	}
	return strings.Join(fields, ", "), nil
}

func (q *Query) compileWhere() (string, error) {
	if expr := q.Condition(); expr != nil {
		compiled, _, err := q.compileExpr(expr)
		if err != nil {
			return "", err
		}
		return compiled, nil
	}
	return "", nil
}

func (q *Query) compileExpr(expr ast.Expr) (string, bool, error) {
	switch typedExpr := expr.(type) {
	case *ast.UnaryExpr:
		compiled, err := q.compileUnaryExpr(typedExpr)
		if err != nil {
			return "", false, err
		}
		return compiled, false, nil
	case *ast.BinaryExpr:
		compiled, err := q.compileBinaryExpr(typedExpr)
		if err != nil {
			return "", false, err
		}
		return compiled, false, nil
	case *ast.Ident:
		compiled, err := q.compileIdent(typedExpr)
		if err != nil {
			return "", false, err
		}
		return compiled, false, nil
	case *ast.Const:
		compiled, err := q.compileConst(typedExpr)
		if err != nil {
			return "", false, err
		}
		return compiled, false, nil
	default:
		return "", false, errors.New("unexepected expression type")
	}
}

func (q *Query) compileUnaryExpr(expr *ast.UnaryExpr) (string, error) {
	var op string
	switch expr.Op {
	case token.NOT:
		op = "not "
	case token.MINUS:
		op = "-"
	default:
		return "", q.unexpect(expr.Op, expr.Pos())
	}

	compiledX, _, err := q.compileExpr(expr.X)
	if err != nil {
		return "", err
	}
	return op + compiledX, nil
}

func (q *Query) compileBinaryExpr(expr *ast.BinaryExpr) (string, error) {
	switch expr.Op {
	case token.AND, token.OR:
		x, ok := expr.X.(*ast.BinaryExpr)
		if !ok {
			return "", q.unexpect(x.Token(), x.Pos())
		}
		y, ok := expr.Y.(*ast.BinaryExpr)
		if !ok {
			return "", q.unexpect(y.Token(), y.Pos())
		}
		compiledX, _, err := q.compileExpr(x)
		if err != nil {
			return "", err
		}
		compiledY, _, err := q.compileExpr(y)
		if err != nil {
			return "", err
		}
		op, compiledX, compiledY, err := q.compileOperator(expr.Op, compiledX, compiledY, false)
		if err != nil {
			return "", err
		}
		if x.Op.Precedence() < expr.Op.Precedence() {
			compiledX = "(" + compiledX + ")"
		}
		if y.Op.Precedence() < expr.Op.Precedence() {
			compiledY = "(" + compiledY + ")"
		}
		return compiledX + " " + op + " " + compiledY, nil
	case token.EQL, token.NEQ:
		var x, y ast.Expr
		x, okX := expr.X.(*ast.ExprList)
		y, okY := expr.Y.(*ast.ExprList)
		if okX != okY && (okX || okY) {
			if okX {
				return q.compileBinaryExprWithExprList(expr.Y, x.(*ast.ExprList), expr.Op)
			}
			return q.compileBinaryExprWithExprList(expr.X, y.(*ast.ExprList), expr.Op)
		}
		x, ok := expr.X.(*ast.Ident)
		if !ok {
			return "", q.unexpect(x.Token(), x.Pos())
		}
		xCol := q.source.Cols.ByName(x.(*ast.Ident).Name)
		if xCol == nil {
			return "", q.notDefined(x.(*ast.Ident).Name, x.Pos())
		}
		compiledX, err := q.compileIdent(x.(*ast.Ident))
		if err != nil {
			return "", err
		}
		var compiledY string
		switch y := expr.Y.(type) {
		case *ast.Const:
			compiledY, err = q.compileConst(y)
			if err != nil {
				return "", err
			}
		case *ast.Ident:
			compiledY, err = q.compileIdent(y)
			if err != nil {
				return "", err
			}
		default:
			return "", q.unexpect(y.Token(), y.Pos())
		}
		if err != nil {
			return "", err
		}
		op, compiledX, compiledY, err := q.compileOperator(expr.Op, compiledX, compiledY, xCol.IsArray)
		if err != nil {
			return "", err
		}
		return compiledX + " " + op + " " + compiledY, nil
	case token.LSS, token.LEQ, token.GTR, token.GEQ, token.LIKE:
		x, ok := expr.X.(*ast.Ident)
		if !ok {
			return "", q.unexpect(x.Token(), x.Pos())
		}
		y, ok := expr.Y.(*ast.Const)
		if !ok {
			return "", q.unexpect(y.Token(), y.Pos())
		}
		colType := q.source.Cols.Type(x.Name)
		if colType == nil {
			return "", q.unexpect(x.Token(), x.Pos())
		}
		if expr.Op == token.LIKE {
			if *colType != source.TypeString {
				return "", q.mustBe(x.Name, "string", "any", x.Pos())
			}
			if t := y.Token(); t != token.STRING {
				return "", q.mustBe(y.Value, "string", t.String(), y.Pos())
			}
		} else {
			switch *colType {
			case source.TypeNumber:
				if t := y.Token(); t != token.INT && t != token.FLOAT {
					return "", q.mustBe(y.Value, "number", t.String(), y.Pos())
				}
			case source.TypeTime:
				if t := y.Token(); t != token.STRING {
					return "", q.mustBe(y.Value, "time", t.String(), y.Pos())
				}
			default:
				return "", q.mustBe(x.Name, "number or time", "any", x.Pos())
			}
		}
		compiledX, err := q.compileIdent(x)
		if err != nil {
			return "", err
		}
		compiledY, err := q.compileConst(y)
		if err != nil {
			return "", err
		}
		op, compiledX, compiledY, err := q.compileOperator(expr.Op, compiledX, compiledY, false)
		if err != nil {
			return "", err
		}
		return compiledX + " " + op + " " + compiledY, nil
	default:
		return "", q.unexpect(expr.Op, expr.Pos())
	}
}

func (q *Query) compileBinaryExprWithExprList(x ast.Expr, y *ast.ExprList, op token.Token) (string, error) {
	typedX, ok := x.(*ast.Ident)
	if !ok {
		return "", q.unexpect(x.Token(), x.Pos())
	}
	column := q.source.Cols.ByName(typedX.Name)
	if column == nil {
		return "", q.notDefined(typedX.Name, x.Pos())
	}
	if isExprArray(y) {
		if column.Type == source.TypeObject {
			return "", q.mustBe(column.Name, "boolean/numeric/text/timestamp", "object", x.Pos())
		}
		if column.IsArray {
			return "", q.mustBe(column.Name, "boolean/numeric/text/timestamp", "not array", x.Pos())
		}
		compiledX, err := q.compileIdent(typedX)
		if err != nil {
			return "", err
		}
		compiledY, err := q.compileExprList(y, column, true)
		if err != nil {
			return "", err
		}
		compiledOp := " in "
		if op == token.NEQ {
			compiledOp = " not in "
		}
		return compiledX + compiledOp + compiledY, nil
	}

	if column.Type != source.TypeObject {
		return "", q.mustBe(column.Name, "array of object", "any", x.Pos())
	}
	compiledY, err := q.compileExprList(y, column, false)
	if err != nil {
		return "", err
	}
	if column.IsArray {
		return "exists (select 1 from (select jsonb_array_elements(" + column.DBName + "::jsonb) item) q where " + compiledY + ")", nil
	}
	return compiledY, nil
}

func isExprArray(expr *ast.ExprList) bool {
	if expr == nil || len(expr.Exprs) == 0 {
		return false
	}
	_, isArray := expr.Exprs[0].(*ast.Const)
	return isArray
}

func (q *Query) compileExprList(expr *ast.ExprList, column *source.Col, isArray bool) (string, error) {
	if expr == nil || len(expr.Exprs) == 0 {
		return "", errors.New("unexpected empty expression list")
	}
	var compiled []string
	for _, el := range expr.Exprs {
		switch typedEl := el.(type) {
		case *ast.Const:
			if !isArray {
				return "", q.unexpect(typedEl.Token(), typedEl.Pos())
			}
			compiledConst, err := q.compileConst(typedEl)
			if err != nil {
				return "", err
			}
			compiled = append(compiled, compiledConst)
		case *ast.UnaryExpr:
			if !isArray {
				return "", q.unexpect(typedEl.Op, typedEl.Pos())
			}
			compiledExpr, err := q.compileUnaryExpr(typedEl)
			if err != nil {
				return "", err
			}
			compiled = append(compiled, compiledExpr)
		case *ast.BinaryExpr:
			var compiledExpr string
			var err error
			if column.IsArray {
				compiledExpr, err = q.compileArrayOfObject(typedEl, column)
			} else {
				compiledExpr, err = q.compileObject(typedEl, column)
			}
			if err != nil {
				return "", err
			}
			compiled = append(compiled, compiledExpr)
		default:
			return "", q.unexpect(typedEl.Token(), typedEl.Pos())
		}
	}

	if isArray {
		return "(" + strings.Join(compiled, ", ") + ")", nil
	}
	return strings.Join(compiled, " and "), nil
}

func (q *Query) compileArrayOfObject(expr *ast.BinaryExpr, column *source.Col) (string, error) {
	ident, ok := expr.X.(*ast.Ident)
	if !ok {
		return "", q.unexpect(expr.X.Token(), expr.X.Pos())
	}
	name := column.Name + "." + ident.Name
	t := q.source.Cols.Type(name)
	if t == nil {
		return "", q.notDefined(name, ident.Pos())
	}
	var typeCast string
	switch *t {
	case source.TypeBool:
		typeCast = "boolean"
	case source.TypeNumber:
		typeCast = "numeric"
	case source.TypeString:
		typeCast = "text"
	case source.TypeTime:
		typeCast = "timestamp"
	default:
		return "", q.mustBe(name, "boolean/numeric/text/timestamp", "any", ident.Pos())
	}
	path, ok := q.source.Cols.JSONPath(name)
	if !ok {
		return "", q.notDefined(name, ident.Pos())
	}
	var compiledY string
	yConst, ok := expr.Y.(*ast.Const)
	needTypeCast := ok
	if !ok {
		yIdent, ok := expr.Y.(*ast.Ident)
		if !ok {
			return "", q.unexpect(expr.Y.Token(), expr.Y.Pos())
		}
		if yIdent.Name != "null" && yIdent.Name != "true" && yIdent.Name != "false" {
			return "", q.unexpect(expr.Y.Token(), expr.Y.Pos())
		}
		compiledY = yIdent.Name
	} else {
		var err error
		compiledY, err = q.compileConst(yConst)
		if err != nil {
			return "", err
		}
	}

	op, _, compiledY, err := q.compileOperator(expr.Op, "", compiledY, false)
	if err != nil {
		return "", err
	}
	if needTypeCast {
		return "(q.item #>> '{" + path + "}')::" + typeCast + " " + op + " " + compiledY + "::" + typeCast, nil
	}
	return "(q.item #>> '{" + path + "}')::" + typeCast + " " + op + " " + compiledY, nil
}

func (q *Query) compileObject(expr *ast.BinaryExpr, column *source.Col) (string, error) {
	ident, ok := expr.X.(*ast.Ident)
	if !ok {
		return "", q.unexpect(expr.X.Token(), expr.X.Pos())
	}
	name := column.Name + "." + ident.Name
	t := q.source.Cols.Type(name)
	if t == nil {
		return "", q.notDefined(name, ident.Pos())
	}
	var typeCast string
	switch *t {
	case source.TypeBool:
		typeCast = "boolean"
	case source.TypeNumber:
		typeCast = "numeric"
	case source.TypeString:
		typeCast = "text"
	case source.TypeTime:
		typeCast = "timestamp"
	default:
		return "", q.mustBe(name, "boolean/numeric/text/timestamp", "any", ident.Pos())
	}
	path, ok := q.source.Cols.JSONPath(name)
	if !ok {
		return "", q.notDefined(name, ident.Pos())
	}
	var compiledY string
	yConst, ok := expr.Y.(*ast.Const)
	if !ok {
		yIdent, ok := expr.Y.(*ast.Ident)
		if !ok {
			return "", q.unexpect(expr.Y.Token(), expr.Y.Pos())
		}
		if yIdent.Name != "null" && yIdent.Name != "true" && yIdent.Name != "false" {
			return "", q.unexpect(expr.Y.Token(), expr.Y.Pos())
		}
		compiledY = yIdent.Name
	} else {
		var err error
		compiledY, err = q.compileConst(yConst)
		if err != nil {
			return "", err
		}
	}

	op, _, compiledY, err := q.compileOperator(expr.Op, "", compiledY, false)
	if err != nil {
		return "", err
	}
	return "(" + column.DBName + " #>> '{" + path + "}')::" + typeCast + " " + op + " " + compiledY + "::" + typeCast, nil
}

func (q *Query) compileIdent(expr *ast.Ident) (string, error) {
	switch expr.Name {
	case "true", "false", "null":
		return expr.Name, nil
	default:
		column := q.source.Cols.ByName(expr.Name)
		if column == nil {
			return "", q.notDefined(expr.Name, expr.Pos())
		}
		return column.DBName, nil
	}
}

func (q *Query) compileConst(expr *ast.Const) (string, error) {
	switch expr.Token() {
	case token.INT, token.FLOAT:
		return expr.Value, nil
	case token.STRING:
		return "'" + expr.Value + "'", nil
	default:
		return "", q.unexpect(expr.Token(), expr.Pos())
	}
}

func (q *Query) compileOperator(op token.Token, x, y string, isArray bool) (string, string, string, error) {
	switch op {
	case token.AND:
		return "and", x, y, nil
	case token.OR:
		return "or", x, y, nil
	case token.EQL:
		if isArray {
			if strings.HasPrefix(y, "'") && strings.HasSuffix(y, "'") {
				return "@>", x, `'"` + y[1:len(y)-1] + `"'`, nil
			}
			return "@>", x, "'" + y + "'", nil
		}
		if y == "null" {
			return "is", x, y, nil
		}
		return "=", x, y, nil
	case token.NEQ:
		if isArray {
			if strings.HasPrefix(y, "'") && strings.HasSuffix(y, "'") {
				return "@>", "not" + x, `'"` + y[1:len(y)-1] + `"'`, nil
			}
			return "@>", "not" + x, "'" + y + "'", nil
		}
		if y == "null" {
			return "is not", x, y, nil
		}
		return "!=", x, y, nil
	case token.LSS:
		return "<", x, y, nil
	case token.LEQ:
		return "<=", x, y, nil
	case token.GTR:
		return ">", x, y, nil
	case token.GEQ:
		return ">=", x, y, nil
	case token.LIKE:
		if !strings.HasPrefix(y, "'") || strings.HasSuffix(y, "'") {
			return "", "", "", fmt.Errorf("like can only be used with strings")
		}
		return "like", x, "'%" + y[1:len(y)-1] + "%'", nil
	default:
		return "", "", "", fmt.Errorf("%v is not operator", op)
	}
}

func (q *Query) compileOrderBy() (string, error) {
	if q.orderBy == nil || len(*q.orderBy) == 0 {
		return "", nil
	}
	orderBy := make([]string, len(*q.orderBy))
	for i, f := range *q.orderBy {
		var compiled string
		column := q.source.Cols.ByName(f.Field.Name)
		if column == nil {
			return "", q.notDefined(f.Field.Name, f.Field.Pos())
		}
		path, ok := q.source.Cols.JSONPath(f.Field.Name)
		if ok {
			parts := strings.Split(f.Field.Name, ".")
			mainColumn := q.source.Cols.ByName(parts[0])
			if mainColumn == nil {
				return "", q.notDefined(parts[0], f.Field.Pos())
			}
			compiled = "(" + mainColumn.DBName + " #>> '{" + path + "}')::" + q.compileType(column.Type)
		} else {
			compiled = column.DBName
		}
		orderBy[i] = compiled + " " + string(f.Direction.Value)
	}
	return strings.Join(orderBy, ", "), nil
}

func (q *Query) compileLimits() (string, error) {
	if q.limits == nil {
		return "", nil
	}
	var limits string
	if q.limits.Len != nil {
		limits += "limit " + q.limits.Len.Value
	}
	if q.limits.From != nil {
		limits += "offset " + q.limits.From.Value
	}
	return limits, nil
}

// return error "unexpected ... at ..."
func (q *Query) unexpect(tok token.Token, pos token.Pos) error {
	return fmt.Errorf("unexpected %v at %v", tok, pos)
}

// return error "... at ... is not defined"
func (q *Query) notDefined(name string, pos token.Pos) error {
	return fmt.Errorf("%v as %v is not defined", name, pos)
}

// return error "... at ... must be ... not ..."
func (q *Query) mustBe(name, expected, got string, pos token.Pos) error {
	return fmt.Errorf("%v at %v must be %v not %v", name, pos, expected, got)
}

func (q *Query) compileType(t source.Datatype) string {
	switch t {
	case source.TypeBool:
		return "boolean"
	case source.TypeNumber:
		return "numeric"
	case source.TypeString:
		return "text"
	case source.TypeTime:
		return "timestamp"
	default:
		return "text"
	}
}
