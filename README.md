# gaddis

Gaddis pseudocode compiler implementation, based on "Starting Out with Programming Logic and Design" by Tony Gaddis

## Install and Use

```bash
~/github/go/src/github.com/dragonsinth/gaddis go install github.com/dragonsinth/gaddis/...
~/github/go/src/github.com/dragonsinth/gaddis gaddis ./examples/2.gad 
What were the annual sales?
integer> 120000
The company made about $27600
```

## Features

- Abstract Syntax Tree (AST)
- Handwritten lexer to tokenize Gaddis pseudocode
- Handwritter parser to build the AST
- Go code generator to translate the AST into Go code
- Out of the box compile-and-execute pseudocode via Go translation and compilation

## Status

Implemented up through Chapter 2; supports basic statements, expressions, operations, variables, constants, Integer, Real, String.
Should cover the whole language by May 2025.

### Not yet supported
- comments
- boolean data type, logic, operators, conditionals
- if statements
- loops
- arrays
- string processing and characters
- functions/modules
- classes
- I/O
  
