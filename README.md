# gaddis

Gaddis pseudocode compiler implementation, based on "Starting Out with Programming Logic and Design" by Tony Gaddis

## Features

- Abstract Syntax Tree (AST)
- Handwritten lexer to tokenize Gaddis pseudocode
- Handwritter parser to build the AST
- Go code generator to translate the AST into Go code
- Out of the box compile-and-execute pseudocode via Go translation and compilation

## Status

Implemented up through Chapter 2; supports basic statements, expressions, operations, variables, constants, literals.
Should cover the whole language by May 2025.

### Not yet supported
- functions/modules
- classes
- I/O
  
