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
	COMMENT

	INTEGER
	REAL
	STRING
	CHARACTER
	BOOLEAN

	IDENT

	ADD
	SUB
	MUL
	DIV
	EXP
	MOD

	EQ
	NEQ
	LT
	LTE
	GT
	GTE
	AND
	OR
	NOT

	ASSIGN
	COMMA
	LPAREN
	RPAREN
	COLON

	CONSTANT
	DECLARE
	DISPLAY
	INPUT
	SET

	END
	IF
	THEN
	ELSE
	SELECT
	CASE
	DEFAULT

	DO
	WHILE
	UNTIL
	FOR
	TO
	STEP

	INT_LIT
	REAL_LIT
	STR_LIT
	CHR_LIT
	TRUE
	FALSE
)

var tokens = []string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	EOL:     "EOL",
	COMMENT: "//",

	INTEGER:   "INTEGER",
	REAL:      "REAL",
	STRING:    "STRING",
	CHARACTER: "CHARACTER",
	BOOLEAN:   "BOOLEAN",
	IDENT:     "IDENT",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",
	EXP: "^",
	MOD: "MOD",

	EQ:  "==",
	NEQ: "!=",
	LT:  "<",
	LTE: "<=",
	GT:  ">",
	GTE: ">=",

	AND: "AND",
	OR:  "OR",
	NOT: "NOT",

	ASSIGN: "=",
	COMMA:  ",",
	LPAREN: "(",
	RPAREN: ")",
	COLON:  ":",

	CONSTANT: "CONSTANT",
	DECLARE:  "DECLARE",
	DISPLAY:  "DISPLAY",
	INPUT:    "INPUT",
	SET:      "SET",

	END:     "END",
	IF:      "IF",
	THEN:    "THEN",
	ELSE:    "ELSE",
	SELECT:  "SELECT",
	CASE:    "CASE",
	DEFAULT: "DEFAULT",

	DO:    "DO",
	WHILE: "WHILE",
	UNTIL: "UNTIL",
	FOR:   "FOR",
	TO:    "TO",
	STEP:  "STEP",

	INT_LIT:  "INT_LIT",
	REAL_LIT: "REAL_LIT",
	STR_LIT:  "STR_LIT",
	CHR_LIT:  "CHR_LIT",
	TRUE:     "TRUE",
	FALSE:    "FALSE",
}

var keywords = map[string]Token{
	"Integer":   INTEGER,
	"Real":      REAL,
	"String":    STRING,
	"Character": CHARACTER,
	"Boolean":   BOOLEAN,
	"MOD":       MOD,
	"AND":       AND,
	"OR":        OR,
	"NOT":       NOT,
	"Constant":  CONSTANT,
	"Declare":   DECLARE,
	"Display":   DISPLAY,
	"Input":     INPUT,
	"Set":       SET,
	"End":       END,
	"If":        IF,
	"Then":      THEN,
	"Else":      ELSE,
	"Select":    SELECT,
	"Case":      CASE,
	"Default":   DEFAULT,
	"Do":        DO,
	"While":     WHILE,
	"Until":     UNTIL,
	"For":       FOR,
	"To":        TO,
	"Step":      STEP,
	"True":      TRUE,
	"False":     FALSE,
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
	s.column++
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
	Pos    int
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
	for !l.stream.Eof() {
		r := l.stream.Peek()
		switch r {
		case '\n':
			return Result{l.advance(), EOL, "\n", nil}
		case '+':
			return Result{l.advance(), ADD, "+", nil}
		case '-':
			pos := l.advance()
			if isDigit(l.stream.Peek()) {
				// Try to parse a numeric literal
				ret := l.parseNumber()
				if ret.Error == nil {
					// back fix the previously consumed negation into the literal
					ret.Pos = pos
					ret.Text = "-" + ret.Text
					return ret
				}
			}
			return Result{pos, SUB, "-", nil}
		case '*':
			return Result{l.advance(), MUL, "*", nil}
		case '/':
			pos := l.advance()
			if l.stream.Peek() == '/' {
				// single line comment
				return l.parseComment(pos)
			}
			return Result{pos, DIV, "/", nil}
		case '^':
			return Result{l.advance(), EXP, "^", nil}
		case '=':
			pos := l.advance()
			if l.stream.Peek() == '=' {
				l.advance()
				return Result{pos, EQ, "==", nil}
			}
			return Result{pos, ASSIGN, "=", nil}
		case '!':
			pos := l.advance()
			if l.stream.Peek() == '=' {
				l.advance()
				return Result{pos, NEQ, "!=", nil}
			}
			return Result{pos, ILLEGAL, string(r), ErrSyntax}
		case '<':
			pos := l.advance()
			if l.stream.Peek() == '=' {
				l.advance()
				return Result{pos, LTE, "<=", nil}
			}
			return Result{pos, LT, "<", nil}
		case '>':
			pos := l.advance()
			if l.stream.Peek() == '=' {
				l.advance()
				return Result{pos, GTE, ">=", nil}
			}
			return Result{pos, GT, ">", nil}
		case ',':
			return Result{l.advance(), COMMA, "=", nil}
		case '(':
			return Result{l.advance(), LPAREN, "(", nil}
		case ')':
			return Result{l.advance(), RPAREN, ")", nil}
		case ':':
			return Result{l.advance(), COLON, ":", nil}
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
	return Result{l.position(), EOF, "", nil}
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
	for !l.stream.Eof() {
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

func (l *Lexer) parseComment(pos Position) Result {
	for !l.stream.Eof() {
		c := l.stream.Next()
		if c == '\n' {
			return Result{pos, EOL, "\n", nil}
		}
	}
	return Result{pos, EOF, "", nil}
}

func (l *Lexer) parseIdent() Result {
	pos := l.position()
	var lit []byte
	for !l.stream.Eof() {
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
	for !l.stream.Eof() {
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
	for !l.stream.Eof() {
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
		Pos:    l.stream.pos,
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
