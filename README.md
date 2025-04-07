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

Implemented up through Chapter 14; supports:

- Basic statements, expressions, operations, variables, constants, control structures.
- `Integer`, `Real`, `String`, `Character`, `Boolean`
- `Input`, `Display`
- `Module` and `Function` declarations
- Arrays, `For Each`
- Indexing `String` as `Character`
- External function library
- File I/O
- Classes

Should cover the whole language by May 2025.

### TODO

- Consider enforcing field vs. record separators in file I/O.

- Consider extending multi-line parsing/printing to all comma-delimited lists:
  - parameter lists
  - argument lists
  - Display/Read/Write

- Classes
  - Enforce `Public` / `Private` in type checking
  - Prevent certain types of class super/sub name collisions?
  - test nil field ref / method call errors
  - implicit/required constructor calls...?
  - Debugger support for classes.

### Not yet supported
- `Delete "filename"`
- `Rename "oldname" "newname"`

## Errata / Differences from the Book

Gaddis Pseudocode is a bit underspecified (on purpose), so I've had to make a few choices here
and there, or fill in gaps. Here's some possible differences (clarifications?) from the book:

- `Character` is an explicit type
  - Character literals are denoted using single-quoted characters: `'X'`
  - (Character type is listed in the reference but not the main text.)

- `String` values are immutable and copy-on-write under the hood. Updating a string in any way
  creates a new `String` and assigns it back into the given reference; no other copies of the
  original `String` are affected.

- All expressions are evaluated left to right, including assignment statements.

- `Display` statements only accept primitive types, not arrays or classes.

- `For` loop variable expression is evaluated once on loop start, and again once per iteration.

- `For` loop stop expression is re-evaluated on every iteration.

- `For` loop step expression must be a numeric constant.
  - This is implied by the book but not explicitly stated.
  - Avoids more complicated code generation where the test expression (`<=` vs `>=`)
    would need to change each iteration depending on the current value of the step expression.
  - Contrast with Basic's `downto` keyword.

- `For Each` loop reference expression is evaluated once per iteration.

- `For Each` loop array expression is evaluated only once when the loop is initialized.

- New library function: `toString()` convert a value of any type to String.
  - (In the book, this conversion only happens when invoking a `Display` statement, but it
    seems like an obvious addition for useful string processing and formatting.)

- Nested `Module` and `Function` declarations are not supported.
  - (This is implied by the book but not explicitly stated.)

- Programs with a `Module main()` are allowed to first execute arbitrary statements in the global block.
  - Global block statements first execute in lexical order, then `main()` is called afterward.

- `Input` statements internally loop until the user inputs a line that can be correctly
  parsed to the type of the input variable; e.g. a non-numeric input will loop and retry
  if the input variable is an `Integer`.
  - The variable expression is only evaluated once, regardless.

- By contrast, `Read` statements in an `InputFile` where the record data is the wrong type
  exit the program immediately with an exception.

- Arrays are filled with zero values on initialization when there is no initializer. When too
  few initializer expressions are provided, the remainder of the array is filled with zero values.
  - (Should too few initializer expressions be an error? It's not clear.)

- Array initializers are currently the _only_ construct that may be parsed across multiple lines.
  - There are explicit examples of this in the book.
  - All other statements always appear on a single line.
  - (Should we allow line breaks across other comma delimited lists like parameter or argument lists?)

- Arrays are deep copied when passed by value.
  - (Implied by the book, but not explicitly stated.)

- Array variables cannot be reassigned.
  - (Implied by the book, but not explicitly stated.)

- Array elements which are themselves Arrays (ie, part of a multidimensional array) cannot be
  reassigned either.
  - For example, giben `Integer table[3][4], row[4]`, you cannot `Set table[0] = someRow`.
  - However, such elements _may_ be passed as arguments, either by value or reference.
  - For example, `Call printValues(table[3])` is legal.
  - (Implied by the book, but not explicitly stated.)
 
- Arrays cannot be the return value of a Function
  - (Implied by the book, but not explicitly stated.)

- Use `Call` to call the external library string modules `insert` and `delete`.
  - This seems like an oversight / misprint? Otherwise these two modules would have a unique syntax just to themselves.

- `Print` outputs to stderr; we didn't implement print support :grin:

- When reading and writing records with multiple fields using file I/O, there is no internal
  distinction between field and record separation. Writing or Reading multiple values in a
  single statement is equivalent to multiple sequential Write or Read statements.
