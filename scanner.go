package w3sql

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

// peek return next character or -1 if that impossible
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

// scanIdentifier return token consisting of latin letters, decimal digits, dots and/or underscores
func (s *Scanner) scanIdentifier() string {
	offs := s.offset

	for isLetter(s.ch) || isDigit(s.ch) || s.ch == '.' || s.ch == '_' {
		s.next()
	}

	return string(s.src[offs:s.offset])
}

// scanNumber return token consisting of decimal digits and/or dot
func (s *Scanner) scanNumber() (token, string) {
	offs := s.offset
	tok := ILLEGAL

	for isDigit(s.ch) || s.ch == '.' {
		if s.ch == '.' {
			tok = FLOAT
		}

		if tok != FLOAT {
			tok = INT
		}

		s.next()
	}

	return tok, string(s.src[offs:s.offset])
}

// scanString return token from quote to quote
func (s *Scanner) scanString() string {
	offs := s.offset
	for s.ch != '"' || s.offset == offs {
		s.next()
	}

	s.next()

	return string(s.src[offs : s.offset-1])
}

// isLetter return true if character is letter or underscore
func isLetter(ch rune) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_'
}

// isLetter return true if character is digit
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// Scan return next token type, his position and literal
func (s *Scanner) scan() (pos int, tok token, lit string) {
	tok = ILLEGAL
	s.skipWhitespace()

	pos = s.offset
	switch ch := s.ch; {
	case isLetter(ch):
		lit = s.scanIdentifier()
		tok = IDENT

	case isDigit(ch) || ch == '.' && isDigit(s.peek()):
		tok, lit = s.scanNumber()

	default:
		s.next()
		switch ch {
		case -1:
			tok = EOF
		case '"':
			tok = STRING
			lit = s.scanString()

		case ':':
			tok = COLON
		case ',':
			tok = COMMA
		case '(':
			tok = LPAREN
		case ')':
			tok = RPAREN
		case '{':
			tok = LBRACE
		case '}':
			tok = RBRACE
		case '[':
			tok = LBRACK
		case ']':
			tok = RBRACK
		case '+':
			tok = ADD
		case '-':
			tok = SUB
		case '*':
			tok = MUL
		case '/':
			tok = QUO
		case '%':
			tok = REM
		case '<':
			if s.ch == '=' {
				s.next()
				tok = LEQ
			} else {
				tok = LSS
			}
		case '>':
			if s.ch == '=' {
				s.next()
				tok = GEQ
			} else {
				tok = GTR
			}
		case '=':
			tok = EQL
		case '!':
			if s.ch == '=' {
				s.next()
				tok = NEQ
			} else {
				tok = NOT
			}
		case '&':
			tok = AND
		case '|':
			tok = OR
		case '@':
			tok = AT
		case '?':
			tok = QUERY
		}
	}

	return
}
