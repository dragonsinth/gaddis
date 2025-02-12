package lex

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
