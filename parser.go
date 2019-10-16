package w3sql

import (
	"fmt"
	"strconv"
)

// A Parser contains a Scanner and current token
type Parser struct {
	scanner Scanner
	pos     int
	tok     token
	lit     string
}

// init initializes a Scanner
func (p *Parser) init(src []rune) {
	p.scanner.Init(src)
}

// next tries get next token
func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.scanner.scan()
}

// return error "unexpected ... at ..."
func (p *Parser) unexpect() error {
	return fmt.Errorf("unexpected %v at %v", p.tok, p.pos)
}

// parseIdent return identifier
func (p *Parser) parseIdent() (*Ident, error) {
	if p.tok != IDENT {
		return nil, p.unexpect()
	}

	return &Ident{Pos: p.pos, Name: p.lit, tok: IDENT}, nil
}

// appendIdent append udent to path
func appendIdent(path *IdentList, ident *Ident) {
	if ident == nil {
		return
	}
	*path = append(*path, ident)
	ident = nil
}

// parsePathAndFields return list of path and fields identifiers
func (p *Parser) parsePathAndFields() (*IdentList, *IdentList, error) {
	path := &IdentList{}
	var fields *IdentList
	var ident *Ident
	var err error
	var prev token

	for {
		p.next()
		switch p.tok {
		case IDENT:
			ident, err = p.parseIdent()
			if err != nil {
				return nil, nil, err
			}

		case QUO:
			appendIdent(path, ident)

		case AT, COMMA:
			if prev == IDENT {
				fields, err = p.parseFields(ident)
			} else {
				fields, err = p.parseFields(nil)
			}
			if err != nil {
				return nil, nil, err
			}

		case QUERY, LBRACK, EOF:
			appendIdent(path, ident)
			return path, fields, nil

		default:
			return nil, nil, p.unexpect()
		}
		prev = p.tok
	}
}

// parsePathAndFields return list of path identifiers
func (p *Parser) parseFields(alreadyIdent *Ident) (*IdentList, error) {
	var fields *IdentList
	var ident *Ident
	var err error

	if alreadyIdent != nil {
		fields = &IdentList{alreadyIdent}
	}

	for p.tok != AT {
		switch p.tok {
		case IDENT:
			ident, err = p.parseIdent()
			if err != nil {
				return nil, err
			}

		case COMMA, EOF:
			appendIdent(fields, ident)

		default:
			return nil, p.unexpect()
		}

		p.next()
	}
	appendIdent(fields, ident)

	return fields, nil
}

// parseLimits return limits info
func (p *Parser) parseLimits() (*Limits, error) {
	lim := &Limits{}

	p.next()
	switch p.tok {
	case COLON:
		lim.From = 0

	case INT:
		from, err := strconv.Atoi(p.lit)
		if err != nil {
			return nil, err
		}
		lim.From = from

		p.next()
		if p.tok != COLON {
			return nil, p.unexpect()
		}

	default:
		return nil, p.unexpect()
	}

	p.next()
	switch p.tok {
	case RBRACK:
		lim.Len = -1

	case INT:
		len, err := strconv.Atoi(p.lit)
		if err != nil {
			return nil, err
		}
		lim.Len = len

		p.next()
		if p.tok != RBRACK {
			return nil, p.unexpect()
		}

	default:
		return nil, p.unexpect()
	}

	return lim, nil
}

// parseExpr return expression
func (p *Parser) parseExpr() (Expr, error) {
	p.next()

	if p.tok == EOF {
		return nil, nil
	}

	x, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	p.next()

	switch {
	case p.tok.IsOperator() && p.tok != COMMA:
		return p.parseBinaryExpr(x)
	case p.tok == COMMA:
		return x, nil
	case p.tok == EOF || p.tok == LBRACK:
		return x, nil
	case p.tok == RBRACE:
		return x, nil
	case p.tok == RPAREN:
		return x, nil

	default:
		return nil, p.unexpect()
	}
}

// parseUnaryExpr return unary expression
func (p *Parser) parseUnaryExpr() (Expr, error) {
	switch p.tok {
	case IDENT:
		ident, err := p.parseIdent()
		if err != nil {
			return nil, err
		}

		return ident, nil

	case INT, FLOAT, STRING:
		return &Const{Pos: p.pos, Value: p.lit, tok: p.tok}, nil

	case LBRACE:
		return p.parseExprList()

	case LPAREN:
		return p.parseExpr()

	case SUB:
		p.next()
		pos := p.pos

		switch p.tok {
		case IDENT:
			ident, err := p.parseIdent()
			if err != nil {
				return nil, err
			}

			return UnaryExpr{Pos: pos, Op: SUB, X: ident}, nil

		case INT, FLOAT:
			return UnaryExpr{Pos: pos, Op: SUB, X: &Const{Pos: p.pos, Value: p.lit, tok: p.tok}}, nil

		case LPAREN:
			expr, err := p.parseExpr()
			if err != nil {
				return nil, err
			}

			return UnaryExpr{Pos: pos, Op: SUB, X: expr}, nil

		default:
			return nil, p.unexpect()
		}
	}

	return nil, nil
}

// parseBinaryExpr return binary expression
func (p *Parser) parseBinaryExpr(x Expr) (Expr, error) {
	expr := BinaryExpr{
		X:   x,
		Op:  p.tok,
		Pos: p.pos,
	}

	y, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	expr.Y = y

	if b, ok := y.(BinaryExpr); ok {
		if b.Op.Precedence() < expr.Op.Precedence() {
			expr.X = BinaryExpr{
				X:  expr.X,
				Op: expr.Op,
				Y:  b.X,
			}
			expr.Op = b.Op
			expr.Y = b.Y
		}
	}

	return expr, nil
}

// parseExprList return expression list
func (p *Parser) parseExprList() (Expr, error) {
	exprList := &ExprList{
		Pos: p.pos,
		tok: LIST,
	}

	for {
		if p.tok == RBRACE {
			break
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}

		exprList.Value = append(exprList.Value, expr)
	}

	return exprList, nil
}

// Parse return ast as Query
func (p *Parser) Parse(s *Server, src string) (*Query, error) {
	p.init([]rune(src))
	q := &Query{server: s}

	path, fields, err := p.parsePathAndFields()
	if err != nil {
		return nil, err
	}
	q.path = path
	q.fields = fields

	if p.tok == EOF {
		return q, nil
	}

	if p.tok == QUERY {
		c, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		q.condition = c
	}

	if p.tok == EOF {
		return q, nil
	}

	if p.tok != LBRACK {
		return nil, p.unexpect()
	}

	lim, err := p.parseLimits()
	if err != nil {
		return nil, err
	}
	q.limits = lim

	p.next()
	if p.tok != EOF {
		return nil, p.unexpect()
	}

	return q, nil
}
