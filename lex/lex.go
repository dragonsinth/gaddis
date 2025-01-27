package lex

import (
	"errors"
	"strconv"
)

type Token int

const (
	ILLEGAL = Token(iota)
	EOF
	EOL

	INTEGER
	REAL
	STRING
	CHARACTER

	IDENT

	ADD
	SUB
	MUL
	DIV
	EXP
	MOD

	ASSIGN
	COMMA
	LPAREN
	RPAREN

	CONSTANT
	DECLARE
	DISPLAY
	INPUT
	SET

	INT_LIT
	REAL_LIT
	STR_LIT
	CHR_LIT
)

var tokens = []string{
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
	EOL:       "EOL",
	INTEGER:   "INTEGER",
	REAL:      "REAL",
	STRING:    "STRING",
	CHARACTER: "CHARACTER",
	IDENT:     "IDENT",
	ADD:       "+",
	SUB:       "-",
	MUL:       "*",
	DIV:       "/",
	EXP:       "^",
	MOD:       "MOD",
	ASSIGN:    "=",
	COMMA:     ",",
	LPAREN:    "(",
	RPAREN:    ")",
	CONSTANT:  "CONSTANT",
	DECLARE:   "DECLARE",
	DISPLAY:   "DISPLAY",
	INPUT:     "INPUT",
	SET:       "SET",
	INT_LIT:   "INT_LIT",
	REAL_LIT:  "REAL_LIT",
	STR_LIT:   "STR_LIT",
	CHR_LIT:   "CHR_LIT",
}

var keywords = map[string]Token{
	"Integer":   INTEGER,
	"Real":      REAL,
	"String":    STRING,
	"Character": CHARACTER,
	"MOD":       MOD,
	"Constant":  CONSTANT,
	"Declare":   DECLARE,
	"Display":   DISPLAY,
	"Input":     INPUT,
	"Set":       SET,
}

func (t Token) String() string {
	return tokens[t]
}

var (
	ErrSyntax                = errors.New("syntax error")
	ErrUnterminatedString    = errors.New("unterminated string literal")
	ErrUnterminatedCharacter = errors.New("unterminated character literal")
	ErrCharacterTooLong      = errors.New("character literal too long")
)

type Stream struct {
	buf    []byte
	pos    int
	line   int
	column int
}

func (s *Stream) Peek() byte {
	return s.buf[s.pos]
}

func (s *Stream) Next() byte {
	ret := s.buf[s.pos]
	s.pos++
	if ret == '\n' {
		s.line++
		s.column = 0
	}
	return ret
}

func (s *Stream) Eof() bool {
	return s.pos >= len(s.buf)
}

type Position struct {
	Line   int
	Column int
}

type Result struct {
	Pos   Position
	Token Token
	Text  string
	Error error
}

type Lexer struct {
	stream *Stream
}

func NewLexer(input []byte) *Lexer {
	return &Lexer{
		stream: &Stream{buf: input, pos: 0, line: 1, column: 0},
	}
}

// Lex scans the input for the next token. It returns the position of the token,
// the token's type, and the literal value.
func (l *Lexer) Lex() Result {
	// keep looping until we return a token
	for {
		if l.stream.Eof() {
			return Result{l.position(), EOF, "", nil}
		}
		r := l.stream.Peek()
		switch r {
		case '\n':
			return Result{l.advance(), EOL, "\n", nil}
		case '+':
			return Result{l.advance(), ADD, "+", nil}
		case '-':
			return Result{l.advance(), SUB, "-", nil}
		case '*':
			return Result{l.advance(), MUL, "*", nil}
		case '/':
			return Result{l.advance(), DIV, "/", nil}
		case '^':
			return Result{l.advance(), EXP, "^", nil}
		case '=':
			return Result{l.advance(), ASSIGN, "=", nil}
		case ',':
			return Result{l.advance(), COMMA, "=", nil}
		case '(':
			return Result{l.advance(), LPAREN, "(", nil}
		case ')':
			return Result{l.advance(), RPAREN, ")", nil}
		case '"':
			return l.parseStringLiteral()
		case '\'':
			return l.parseCharacterLiteral()
		default:
			if isSpace(r) {
				l.advance()
				continue
			} else if isDigit(r) {
				return l.parseNumber()
			} else if isIdentStart(r) {
				return l.parseIdent()
			} else {
				return Result{l.position(), ILLEGAL, string(r), ErrSyntax}
			}
		}
	}
}

func (l *Lexer) advance() Position {
	pos := l.position()
	l.stream.Next()
	return pos
}

func (l *Lexer) parseNumber() Result {
	pos := l.position()
	var lit []byte
	isDecimal := false
	for {
		c := l.stream.Peek()
		if isDigit(c) {
			lit = append(lit, c)
			l.advance()
		} else if c == '.' {
			lit = append(lit, c)
			l.advance()
			isDecimal = true
		} else {
			break
		}
	}

	str := string(lit)
	if isDecimal {
		if _, err := strconv.ParseFloat(str, 64); err == nil {
			return Result{pos, REAL_LIT, str, nil}
		} else {
			return Result{pos, ILLEGAL, str, err}
		}
	} else {
		if _, err := strconv.ParseInt(str, 0, 64); err == nil {
			return Result{pos, INT_LIT, str, nil}
		} else {
			return Result{pos, ILLEGAL, str, err}
		}
	}
}

func (l *Lexer) parseIdent() Result {
	pos := l.position()
	var lit []byte
	for {
		c := l.stream.Peek()
		if isIdentCharacter(c) {
			lit = append(lit, c)
			l.advance()
		} else {
			break
		}
	}

	str := string(lit)
	if tok, ok := keywords[str]; ok {
		return Result{pos, tok, str, nil}
	}
	return Result{pos, IDENT, str, nil}
}

func (l *Lexer) parseStringLiteral() Result {
	pos := l.position()
	lit := []byte{'"'}
	l.advance()

	// FIXME: real string literal parse with escaping?
	for {
		c := l.stream.Peek()
		if c == '\n' {
			return Result{pos, ILLEGAL, string(lit), ErrUnterminatedString}
		}
		l.advance()
		lit = append(lit, c)
		if c == '"' {
			break
		}
	}
	return Result{pos, STR_LIT, string(lit), nil}
}

func (l *Lexer) parseCharacterLiteral() Result {
	pos := l.position()
	lit := []byte{'\''}
	l.advance()
	for {
		c := l.stream.Peek()
		if c == '\n' {
			return Result{pos, ILLEGAL, string(lit), ErrUnterminatedCharacter}
		}
		l.advance()
		lit = append(lit, c)
		if c == '\'' {
			break
		}
	}
	// FIXME: real character literal parse with escaping?
	if len(lit) != 3 {
		return Result{pos, ILLEGAL, string(lit), ErrCharacterTooLong}
	}
	return Result{pos, CHR_LIT, string(lit), nil}
}

func (l *Lexer) position() Position {
	return Position{
		Line:   l.stream.line,
		Column: l.stream.column,
	}
}

func isSpace(r byte) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

func isDigit(r byte) bool {
	return r >= '0' && r <= '9'
}

func isIdentStart(r byte) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentCharacter(r byte) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}
