# gaddis

Gaddis pseudocode compiler implementation, based on "Starting Out with Programming Logic and Design" by Tony Gaddis

## Features

### Compiler and Runtime

- Abstract Syntax Tree (AST)
- Handwritten lexer to tokenize Gaddis pseudocode
- Handwritter parser to build the AST
- Symbol resolution, type checking, control flow validation
- Psuedo assembly language and interpreter
- Go code generator to transliterate the AST into Go code, allowing native execution.

### VSCode extension

- Syntax highlighting
- Autoformat
- Inline compile errors
- Debug Adapter Protocol (DAP) debugger

## Install and Use

### VScode (recommended)

- Download `gaddis-vscode.vsix` from the [latest release](https://github.com/dragonsinth/gaddis/releases/latest).
- VSCode -> Settings -> Extensions-> `...`
- `Install from VSIX...`

### Command line gaddis

(See `gaddis help` for additional commands.)

#### Install
```bash
go install github.com/dragonsinth/gaddis/...
```

#### Run

Runs the given file interactively.

```bash
gaddis run ./examples/chapter2/2.gad
```

```
What were the annual sales?
integer> 120000
The company made about $27600
```

#### Test

Runs the given file as a test, using `2.gad.in` as program input,
and `2.gad.out` as the expected test output.

```bash
gaddis test ./examples/chapter2/2.gad
```

```
PASSED
```

NOTE: the first time `test` is run on a file, if input and output files
do not yet exist, `gaddis test` will run in "capture" mode, potentially
reading from stdin to create input and output files for subsequent test runs.

## Status

Implemented up through Chapter 8; supports:

- Basic statements, expressions, operations, variables, constants, control structures.
- `Integer`, `Real`, `String`, `Character`, `Boolean`
- `Module` and `Function` declarations
- `Input`, `Display`, external function library
- Arrays, `For Each`
- Indexing `String` as `Character`

Should cover the whole language by May 2025.

### Not yet supported

- some string processing functions
- file I/O
- classes

## Errata / Differences from the Book

Gaddis Pseudocode is a bit underspecified, so I've had to make a few design choices here
and there, or fill in gaps. Here's some possible differences (or clarifications) from the book:

- Explicit `Character` type and character literals, using single-quoted characters: `'X'`

- For loop variables must be simple variable references, not field or array element references.
    - This is never specified in the book, but there are no counter examples.
    - Avoids unspecified potential re-evaluations of complex reference expressions.

- New lib funcs to convert number to String in an expression context.
  - `String integerToString(n Integer)`
  - `String realToString(n Real)`
  - (In the book, this only happens when invoking a `Display` statement)

- `Print` is just an alias for `Display`, there is no printer support.

- Nested `Module` and `Function` declarations are not currently supported, but could be.
  - This is never specified in the book, but there are no examples of nested functions.

- `Input` statements will loop until the user inputs a line that can be correctly
  parsed to the type of the input variable; e.g. a non-numeric input will loop and retry
  if the input variable is an `Integer`.

- By contrast, `Read` statements in an `InputFile` where the record data is the wrong type
  exit the program immediately with an exception.
