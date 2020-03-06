package scanner

import "github.com/x-foby/w3sql/token"

// A Scanner contains the processed data.
// Before use, it must be initialized using the Init method.
type Scanner struct {
	src    []rune
	len    int
	ch     rune
	offset int
}

// Init sets correct state for a Scanner and tries read first character
func (s *Scanner) Init(src []rune) {
	s.src = src
	s.len = len(s.src)
	s.offset = -1
	s.next()
}

// next increment current offset and sets appropriate current character or -1 if that impossible
func (s *Scanner) next() {
	if s.offset < s.len-1 {
		s.offset++
		s.ch = s.src[s.offset]
	} else {
		s.offset = s.len
		s.ch = -1 // eof
	}
}

// peek returns next character or -1 if that impossible
func (s *Scanner) peek() rune {
	if s.offset < s.len-1 {
		return s.src[s.offset+1]
	}
	return -1 // eof
}

// skipWhitespace increment current offset while current character is whitespace
func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' {
		s.next()
	}
}

// scanIdentifier returnы token.Token consisting of latin letters, decimal digits, dots and/or underscores
func (s *Scanner) scanIdentifier() string {
	offs := s.offset
	for isLetter(s.ch) || isDigit(s.ch) || s.ch == '.' || s.ch == '_' {
		s.next()
	}
	return string(s.src[offs:s.offset])
}

// scanNumber return token.Token consisting of decimal digits and/or dot
func (s *Scanner) scanNumber() (token.Token, string) {
	offs := s.offset
	tok := token.ILLEGAL
	for isDigit(s.ch) || (s.ch == '.' && tok != token.FLOAT) {
		if s.ch == '.' {
			tok = token.FLOAT
		}
		if tok == token.ILLEGAL {
			tok = token.INT
		}
		s.next()
	}
	return tok, string(s.src[offs:s.offset])
}

// scanString return token.Token from quote to quote
func (s *Scanner) scanString() (token.Pos, token.Token, string) {
	offs := s.offset
	for s.ch != '"' || s.offset == offs {
		if s.ch == -1 {
			return token.Pos(s.offset), token.ILLEGAL, ""
		}
		s.next()
	}
	return token.Pos(offs), token.STRING, string(s.src[offs+1 : s.offset])
}

// isLetter return true if character is letter or underscore
func isLetter(ch rune) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_'
}

// isLetter return true if character is digit
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// Scan returns next token.Token type, his position and literal
func (s *Scanner) Scan() (pos token.Pos, tok token.Token, lit string) {
	tok = token.ILLEGAL
	s.skipWhitespace()

	pos = token.Pos(s.offset)
	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanIdentifier()
		tok = token.IDENT
	case isDigit(ch) || ch == '.' && isDigit(s.peek()):
		tok, lit = s.scanNumber()
	default:
		switch ch {
		case -1:
			tok = token.EOF
		case '$':
			if isLetter(s.peek()) {
				s.next()
				if lit = s.scanIdentifier(); lit != "" {
					tok = token.PSEUDO
				}
			} else {
				pos = token.Pos(s.offset + 1)
			}
		case '"':
			pos, tok, lit = s.scanString()
		case ':':
			tok = token.COLON
		case ',':
			tok = token.COMMA
		case '(':
			tok = token.LPAREN
		case ')':
			tok = token.RPAREN
		case '{':
			tok = token.LBRACE
		case '}':
			tok = token.RBRACE
		case '[':
			tok = token.LBRACK
		case ']':
			tok = token.RBRACK
		case '<':
			if s.peek() == '=' {
				s.next()
				tok = token.LEQ
			} else {
				tok = token.LSS
			}
		case '>':
			if s.peek() == '=' {
				s.next()
				tok = token.GEQ
			} else {
				tok = token.GTR
			}
		case '=':
			tok = token.EQL
		case '!':
			if s.peek() == '=' {
				s.next()
				tok = token.NEQ
			} else {
				tok = token.NOT
			}
		case '&':
			tok = token.AND
		case '|':
			tok = token.OR
		case '@':
			tok = token.AT
		case '?':
			tok = token.QUERY
		case '/':
			tok = token.QUO
		case '+':
			tok = token.PLUS
		case '-':
			tok = token.MINUS
		}
		s.next()
	}
	return
}
