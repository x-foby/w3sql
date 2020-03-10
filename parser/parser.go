package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/x-foby/w3sql/ast"
	"github.com/x-foby/w3sql/query"
	"github.com/x-foby/w3sql/scanner"
	"github.com/x-foby/w3sql/token"
)

// A Parser contains a Scanner and current token
type Parser struct {
	scanner scanner.Scanner
	pos     token.Pos
	tok     token.Token
	lit     string
	globals map[string]ast.Expr
}

// New returns new Parser
func New() *Parser {
	return &Parser{
		globals: make(map[string]ast.Expr),
	}
}

// Parse return a Query
func (p *Parser) Parse( /*s *Server, */ src string) (*query.Query, error) {
	p.scanner.Init([]rune(src))

	var (
		path    string
		fields  *ast.IdentList
		expr    ast.Expr
		orderBy *ast.OrderByStmtList
		limits  *ast.LimitsStmt
		err     error
	)

	path, fields, err = p.parsePathAndFields()
	if err != nil {
		return nil, err
	}
	if p.tok == token.QUERY {
		expr, _, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	if p.tok == token.COLON {
		orderBy, err = p.parseOrderByStmt()
		if err != nil {
			return nil, err
		}
	}
	if p.tok == token.LBRACK {
		limits, err = p.parseLimits()
		if err != nil {
			return nil, err
		}
	}
	if p.tok != token.EOF {
		return nil, p.unexpect()
	}
	return query.New(path, fields, expr, orderBy, limits), nil
}

// WithGlobals add global idents to context
func (p *Parser) WithGlobals(globals map[string]ast.Expr) *Parser {
	p.globals = globals
	return p
}

func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.scanner.Scan()
}

// return error "unexpected ... at ..."
func (p *Parser) unexpect() error {
	return fmt.Errorf("unexpected %v at %v", p.tok, p.pos)
}

// parseIdent return identifier
func (p *Parser) parseIdent() (ast.Expr, error) {
	if p.tok != token.IDENT {
		return nil, p.unexpect()
	}
	if global, ok := p.globals[p.lit]; ok {
		return global, nil
	}
	return ast.NewIdent(p.lit, p.pos), nil
}

// parsePathAndFields return list of path identifiers
func (p *Parser) parseFields(alreadyIdent *ast.Ident) (*ast.IdentList, error) {
	var ident *ast.Ident
	fields := ast.NewIdentList()
	if alreadyIdent != nil {
		fields.Append(alreadyIdent)
	}

	for p.tok != token.AT {
		switch p.tok {
		case token.IDENT:
			expr, err := p.parseIdent()
			if err != nil {
				return nil, err
			}
			var ok bool
			ident, ok = expr.(*ast.Ident)
			if !ok {
				return nil, p.unexpect()
			}
		case token.COMMA, token.EOF:
			if ident != nil {
				fields.Append(ident)
			}
		default:
			return nil, p.unexpect()
		}
		p.next()
	}
	if ident != nil {
		fields.Append(ident)
	}
	if len(*fields) == 0 {
		return nil, nil
	}
	return fields, nil
}

// parsePathAndFields return list of path and fields identifiers
func (p *Parser) parsePathAndFields() (string, *ast.IdentList, error) {
	var path []string
	var fields *ast.IdentList
	var ident *ast.Ident
	var err error
	var prev token.Token

	for {
		p.next()
		switch p.tok {
		case token.IDENT:
			expr, err := p.parseIdent()
			if err != nil {
				return "", nil, err
			}
			var ok bool
			ident, ok = expr.(*ast.Ident)
			if !ok {
				return "", nil, p.unexpect()
			}
		case token.QUO:
			if ident != nil {
				path = append(path, ident.Name)
				ident = nil
			}
		case token.AT, token.COMMA:
			if prev == token.IDENT && ident != nil {
				fields, err = p.parseFields(ident)
			} else {
				fields, err = p.parseFields(nil)
			}
			if err != nil {
				return "", nil, err
			}

		case token.QUERY, token.LBRACK, token.COLON, token.EOF:
			if ident != nil {
				path = append(path, ident.Name)
				ident = nil
			}
			return strings.Join(path, "/"), fields, nil

		default:
			return "", nil, p.unexpect()
		}
		prev = p.tok
	}
}

func (p *Parser) parseExpr() (ast.Expr, bool, error) {
	p.next()
	if p.tok == token.EOF {
		return nil, false, nil
	}
	x, isIsolated, err := p.parseUnaryExpr()
	if err != nil {
		return nil, false, err
	}
	p.next()
	if p.tok.IsOperator() {
		expr, _, err := p.parseBinaryExpr(x)
		return expr, isIsolated, err
	}
	switch p.tok {
	case token.COMMA, token.RBRACE, token.RPAREN, token.COLON, token.LBRACK, token.EOF:
		return x, isIsolated, nil
	default:
		return nil, false, p.unexpect()
	}
}

func (p *Parser) parseUnaryExpr() (ast.Expr, bool, error) {
	switch p.tok {
	case token.IDENT:
		expr, err := p.parseIdent()
		return expr, false, err
	case token.INT, token.FLOAT, token.STRING:
		return ast.NewConst(p.lit, p.pos, p.tok), false, nil
	case token.LBRACE:
		expr, err := p.parseExprList()
		return expr, false, err
	case token.LPAREN:
		expr, _, err := p.parseExpr()
		if err != nil {
			return nil, false, err
		}
		return expr, true, nil
	case token.MINUS, token.NOT:
		unary := ast.NewUnaryExpr(p.tok, nil, p.pos)
		p.next()
		expr, _, err := p.parseUnaryExpr()
		if err != nil {
			return nil, false, err
		}
		unary.X = expr
		return unary, false, nil
	default:
		return nil, false, p.unexpect()
	}
}

func (p *Parser) parseBinaryExpr(x ast.Expr) (ast.Expr, bool, error) {
	expr := ast.NewBinaryExpr(p.tok, x, nil, p.pos)
	y, isIsolated, err := p.parseExpr()
	if err != nil {
		return nil, false, err
	}
	if !isIsolated && needSwap(expr, y) {
		swap(expr, y.(*ast.BinaryExpr))
	} else {
		expr.Y = y
	}
	return expr, isIsolated, nil
}

func (p *Parser) parseExprList() (ast.Expr, error) {
	exprList := ast.NewExprList(p.pos)
	for {
		if p.tok == token.RBRACE {
			break
		}
		expr, _, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		exprList.Append(expr)
	}
	return exprList, nil
}

func (p *Parser) parseLimits() (*ast.LimitsStmt, error) {
	var from, length *ast.Const

	p.next()
	switch p.tok {
	case token.COLON:
		from = nil
	case token.INT:
		_, err := strconv.Atoi(p.lit)
		if err != nil {
			return nil, err
		}
		from = ast.NewConst(p.lit, p.pos, token.INT)

		p.next()
		if p.tok != token.COLON {
			return nil, p.unexpect()
		}

	default:
		return nil, p.unexpect()
	}

	p.next()
	switch p.tok {
	case token.RBRACK:
		length = nil
	case token.INT:
		_, err := strconv.Atoi(p.lit)
		if err != nil {
			return nil, err
		}
		length = ast.NewConst(p.lit, p.pos, token.INT)

		p.next()
		if p.tok != token.RBRACK {
			return nil, p.unexpect()
		}
	default:
		return nil, p.unexpect()
	}
	p.next()

	return ast.NewLimitsStmt(from, length), nil
}

func needSwap(x *ast.BinaryExpr, y ast.Expr) bool {
	yBinaryExpr, ok := y.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	return x.Op.Precedence() > yBinaryExpr.Op.Precedence()
}

func swap(x, y *ast.BinaryExpr) {
	*x = *ast.NewBinaryExpr(y.Op, ast.NewBinaryExpr(x.Op, x.X, y.X, x.Pos()), y.Y, y.Pos())

	xx := x.X.(*ast.BinaryExpr)
	if needSwap(xx, xx.Y) {
		swap(xx, xx.Y.(*ast.BinaryExpr))
	}
}

func (p *Parser) parseOrderByStmt() (*ast.OrderByStmtList, error) {
	p.next()
	orderBy := ast.NewOrderByStmtList()
	for {
		switch p.tok {
		case token.LBRACK, token.EOF:
			return orderBy, nil
		case token.COMMA:
			p.next()
			continue
		case token.PLUS, token.MINUS:
			var dir *ast.OrderByDir
			if p.tok == token.PLUS {
				dir = ast.NewOrderByDir(ast.OrderAsc, p.pos, p.tok)
			} else {
				dir = ast.NewOrderByDir(ast.OrderDesc, p.pos, p.tok)
			}
			p.next()
			if p.tok != token.IDENT {
				return nil, p.unexpect()
			}
			orderBy.Append(ast.NewOrderByStmt(ast.NewIdent(p.lit, p.pos), dir))
		default:
			return nil, p.unexpect()
		}
		p.next()
	}
}
