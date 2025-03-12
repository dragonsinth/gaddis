# gaddis

Gaddis pseudocode compiler implementation, based on "Starting Out with Programming Logic and Design" by Tony Gaddis

## Install and Use

### VScode (recommended)

- Download `gaddis-vscode.vsix` from the [latest release](https://github.com/dragonsinth/gaddis/releases/latest).
- VSCode -> Settings -> Extensions-> `...`
- `Install from VSIX...`

### Command line gaddis

```bash
~/github/go/src/github.com/dragonsinth/gaddis go install github.com/dragonsinth/gaddis/...
~/github/go/src/github.com/dragonsinth/gaddis gaddis run ./examples/2.gad 
What were the annual sales?
integer> 120000
The company made about $27600
```

## Features

### Compiler and Runtime
- Abstract Syntax Tree (AST)
- Handwritten lexer to tokenize Gaddis pseudocode
- Handwritter parser to build the AST
- Symbol resolution, type checking, control flow validation
- Psuedo assembly language and interpreter
- Deprecated: Go code generator to transliterate the AST into Go code.

### VSCode extension
- Syntax highlighting
- Autoformat
- Inline compile errors
- Debug Adapter Protocol (DAP) debugger

## Status

Implemented up through Chapter 6; supports:

- Basic statements, expressions, operations, variables, constants
- Integer, Real, String, Character Boolean
- Modules and Functions
- Input, Display, external function library

Should cover the whole language by May 2025.

### Not yet supported

- arrays
- classes
- string processing and characters
- file I/O
