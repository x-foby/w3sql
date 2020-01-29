package w3sql

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// SPACE is space
const SPACE = " "

func compile(q Query, target string) (string, error) {
	if q.path == nil {
		return "", errors.New("path is not defined")
	}

	source, ok := q.server.sources[q.Path()]
	if !ok || source == nil {
		return "", errors.New("source is not defined")
	}

	var buf strings.Builder

	// SELECT
	if err := compileSelect(q.fields, source, &buf); err != nil {
		return "", err
	}

	// FROM
	buf.WriteString("from " + target + SPACE)

	// WHERE
	var where strings.Builder
	appendStr, err := compileWhere(q.condition, source, &where, false)
	if err != nil {
		return "", err
	}

	buf.WriteString(appendStr)

	if where.Len() != 0 {
		buf.WriteString("where " + where.String() + SPACE)
	}

	// LIMITS
	if q.limits != nil {
		if q.limits.Len != -1 {
			buf.WriteString("limit " + strconv.Itoa(q.limits.Len) + SPACE)
		}

		buf.WriteString("offset " + strconv.Itoa(q.limits.From) + SPACE)
	}

	return buf.String(), nil
}

func compileSelect(fields *IdentList, source *Source, buf *strings.Builder) error {
	cols := []string{"*"}
	if fields != nil {
		l := len(*fields)
		cols = make([]string, l)
		for i, field := range *fields {
			col, fields, err := getCol(field.Name, source)
			if err != nil {
				return err
			}
			if len(fields) > 0 {
				return errors.New("can not return field of JSON-object")
			}

			cols[i] = col.Name
		}
	}

	if len(cols) == 1 {
		buf.WriteString("select " + cols[0] + SPACE)
	} else {
		buf.WriteString("select" + SPACE + strings.Join(cols, ","+SPACE) + SPACE)
	}

	return nil
}

func compileWhere(cond Expr, source *Source, buf *strings.Builder, compareWithJSON bool) (string, error) {
	if cond == nil {
		return "", nil
	}

	var appendStr string
	switch t := cond.(type) {
	case *Const:
		if t.tok == STRING {
			if compareWithJSON {
				buf.WriteString(`"` + t.Value + `"`)
			} else {
				buf.WriteString("'" + strings.ReplaceAll(t.Value, "'", "''") + "'")
			}
		} else {
			buf.WriteString(t.Value)
		}

	case *Ident:
		if t.Name == "true" || t.Name == "false" {
			buf.WriteString(t.Name)
			return "", nil
		}

		col, fields, err := getCol(t.Name, source)
		if err != nil {
			return "", nil
		}

		switch col.Type {
		case JSONObject:
			buf.WriteString(`"` + col.Name + `"#>>'{` + strings.Join(fields, ",") + "}'")
		case JSONArray:
			if col.ChildrenType == JSONObject {
				buf.WriteString(`exists(select 1 from jsonb_array_elements("` + col.Name + `") item where item#>>'{` + strings.Join(fields, ",") + "}'")
				appendStr = " limit 1)"
			} else {
				buf.WriteString(`"` + col.Name + `"`)
			}

		default:
			buf.WriteString(`"` + col.Name + `"`)
		}

	case *ExprList:
		l := len(t.Value)
		for i, expr := range t.Value {
			appendStr, err := compileWhere(expr, source, buf, false)
			if err != nil {
				return "", nil
			}

			buf.WriteString(appendStr)

			if i < l-1 {
				buf.WriteString(",")
			}
		}
		// buf.WriteString(t.Op.String())
		// return compileWhere(t.X, source, buf)

	case UnaryExpr:
		buf.WriteString(t.Op.String())
		return compileWhere(t.X, source, buf, false)

	case BinaryExpr:
		if err := compileBinaryExpr(t, source, buf); err != nil {
			return "", nil
		}
	default:
		return "", fmt.Errorf("%s is not defined", t)
	}

	return appendStr, nil
}

func compileBinaryExpr(expr BinaryExpr, source *Source, buf *strings.Builder) error {
	lp, rp := "", ""
	if be, ok := expr.X.(BinaryExpr); ok {
		if be.Op.Precedence() < expr.Op.Precedence() {
			lp, rp = "(", ")"
		}
	}

	buf.WriteString(lp)
	appendStrX, err := compileWhere(expr.X, source, buf, false)
	if err != nil {
		return err
	}
	buf.WriteString(rp)

	isJSONArray := false
	isNEQ := false
	isList := false

	switch expr.Op {
	case AND:
		buf.WriteString(" and ")
	case OR:
		buf.WriteString(" or ")
	case EQL, NEQ:
		if _, ok := expr.Y.(*ExprList); ok {
			isList = true
			if expr.Op == NEQ {
				buf.WriteString(" not")
			}
			buf.WriteString(" in ")
		} else {
			switch t := expr.X.(type) {
			case *Ident:
				col, _, err := getCol(t.Name, source)
				if err != nil {
					return err
				}
				if col.Type == JSONArray {
					buf.WriteString("@>")
					isJSONArray = true
					isNEQ = expr.Op == NEQ
				} else {
					buf.WriteString(expr.Op.String())
				}

			default:
				buf.WriteString(expr.Op.String())
			}
		}

	default:
		buf.WriteString(" " + expr.Op.String() + " ")
	}

	lp, rp = "", ""
	if be, ok := expr.Y.(BinaryExpr); ok {
		if be.Op.Precedence() < expr.Op.Precedence() {
			lp, rp = "(", ")"
		}
	}
	if isList {
		lp, rp = "(", ")"
	}
	qt := ""
	if isJSONArray {
		qt = "'"
	}

	buf.WriteString(qt)
	buf.WriteString(lp)
	appendStrY, err := compileWhere(expr.Y, source, buf, isJSONArray)
	if err != nil {
		return err
	}
	buf.WriteString(appendStrY)
	if appendStrX == " limit 1)" {
		buf.WriteString("::text")
	}
	buf.WriteString(appendStrX)
	buf.WriteString(rp)
	buf.WriteString(qt)

	if isNEQ {
		buf.WriteString(" = false")
	}

	return nil
}

func getCol(name string, source *Source) (*Col, []string, error) {
	col, ok := source.Cols[name]
	if ok {
		return &col, nil, nil
	}

	names := strings.Split(name, ".")
	if len(names) == 1 {
		return nil, nil, fmt.Errorf("%s is not defined", name)
	}

	col, ok = source.Cols[names[0]]
	if !ok {
		return nil, nil, fmt.Errorf("%s is not defined", name)
	}

	if col.Type != JSONObject && col.Type != JSONArray {
		return nil, nil, fmt.Errorf("%s is not JSON", name)
	}

	return &col, names[1:], nil
}
