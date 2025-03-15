package lex

import (
	"errors"
	"strconv"
)

var (
	ErrSyntax                = errors.New("syntax error")
	ErrUnterminatedString    = errors.New("unterminated string literal")
	ErrUnterminatedCharacter = errors.New("unterminated character literal")
	ErrCharacterTooLong      = errors.New("character literal too long")
)

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

func New(input string) *Lexer {
	return &Lexer{
		stream: &Stream{buf: input, pos: 0, line: 0, column: 0},
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
		case '[':
			return Result{l.advance(), LBRACKET, "[", nil}
		case ']':
			return Result{l.advance(), RBRACKET, "]", nil}
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
				return Result{l.advance(), ILLEGAL, string(r), ErrSyntax}
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

	// FIXME: finish numeric parsing and include value with lex?
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
	comment := []byte{'/'}
	for !l.stream.Eof() {
		c := l.stream.Peek()
		if c == '\n' {
			break
		}
		comment = append(comment, l.stream.Next())
	}
	return Result{pos, COMMENT, string(comment), nil}
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
