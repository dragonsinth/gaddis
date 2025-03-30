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
	REF

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
	LBRACKET
	RBRACKET
	COLON
	DOT

	CONSTANT
	DECLARE
	DISPLAY
	PRINT
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
	EACH
	IN

	MODULE
	CALL
	FUNCTION
	RETURN

	INT_LIT
	REAL_LIT
	STR_LIT
	CHR_LIT
	TAB_LIT

	TRUE
	FALSE

	READ
	WRITE
	OPEN
	CLOSE
	OUTPUTFILE
	INPUTFILE
	APPENDMODE

	CLASS
	EXTENDS
	PUBLIC
	PRIVATE
	NEW
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
	REF:       "REF",

	IDENT: "IDENT",

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

	ASSIGN:   "=",
	COMMA:    ",",
	LPAREN:   "(",
	RPAREN:   ")",
	LBRACKET: "(",
	RBRACKET: ")",
	COLON:    ":",
	DOT:      ".",

	CONSTANT: "CONSTANT",
	DECLARE:  "DECLARE",
	DISPLAY:  "DISPLAY",
	PRINT:    "PRINT",
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
	EACH:  "EACH",
	IN:    "IN",

	MODULE:   "MODULE",
	CALL:     "CALL",
	FUNCTION: "FUNCTION",
	RETURN:   "RETURN",

	INT_LIT:  "INT_LIT",
	REAL_LIT: "REAL_LIT",
	STR_LIT:  "STR_LIT",
	CHR_LIT:  "CHR_LIT",
	TAB_LIT:  "TAB_LIT",

	TRUE:  "TRUE",
	FALSE: "FALSE",

	READ:       "READ",
	WRITE:      "WRITE",
	OPEN:       "OPEN",
	CLOSE:      "CLOSE",
	OUTPUTFILE: "OUTPUTFILE",
	INPUTFILE:  "INPUTFILE",
	APPENDMODE: "APPENDMODE",

	CLASS:   "CLASS",
	EXTENDS: "EXTENDS",
	PUBLIC:  "PUBLIC",
	PRIVATE: "PRIVATE",
	NEW:     "NEW",
}

var keywords = map[string]Token{
	"Integer":    INTEGER,
	"Real":       REAL,
	"String":     STRING,
	"Character":  CHARACTER,
	"Boolean":    BOOLEAN,
	"Ref":        REF,
	"MOD":        MOD,
	"AND":        AND,
	"OR":         OR,
	"NOT":        NOT,
	"Constant":   CONSTANT,
	"Declare":    DECLARE,
	"Display":    DISPLAY,
	"Print":      PRINT,
	"Input":      INPUT,
	"Set":        SET,
	"End":        END,
	"If":         IF,
	"Then":       THEN,
	"Else":       ELSE,
	"Select":     SELECT,
	"Case":       CASE,
	"Default":    DEFAULT,
	"Do":         DO,
	"While":      WHILE,
	"Until":      UNTIL,
	"For":        FOR,
	"To":         TO,
	"Step":       STEP,
	"Each":       EACH,
	"In":         IN,
	"Module":     MODULE,
	"Call":       CALL,
	"Function":   FUNCTION,
	"Return":     RETURN,
	"Tab":        TAB_LIT,
	"True":       TRUE,
	"False":      FALSE,
	"Read":       READ,
	"Write":      WRITE,
	"Open":       OPEN,
	"Close":      CLOSE,
	"OutputFile": OUTPUTFILE,
	"InputFile":  INPUTFILE,
	"AppendMode": APPENDMODE,
	"Class":      CLASS,
	"Extends":    EXTENDS,
	"Public":     PUBLIC,
	"Private":    PRIVATE,
	"New":        NEW,
}

func (t Token) String() string {
	return tokens[t]
}
