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

- Resolve questions around type -> string conversions:
  - `String characterToString(Character c)`?
  - `String booleanToString(Boolean b)`?
  - Or should append(s, ?) take any type?

- Consider enforcing field vs. record separators in file I/O.

- Consider extending multi-line parsing/printing to all comma-delimited lists:
  - parameter lists
  - argument lists
  - Display/Read/Write

- Classes
  - Enforce `Public` / `Private` in type checking
  - Prevent certain types of class super/sub name collisions?
  - Class field initializers
    - Zero-initialize new objects in asmgen.
    - Zero-initialize fields containing arrays in gogen.
  - test nil field ref / method call errors
  - default constructor...?
  - implicit/required constructor calls?
  - Debugger support for classes.

### Not yet supported
- `Delete "filename"`
- `Rename "oldname" "newname"`

## Errata / Differences from the Book

Gaddis Pseudocode is a bit underspecified, so I've had to make a few design choices here
and there, or fill in gaps. Here's some possible differences (or clarifications) from the book:

- `Character` is an explicit type; character literals are denoted using single-quoted characters: `'X'`
  - The book only mentions an explicit type in the index without examples; character literals are never defined.

- `String` values are immutable and copy-on-write under the hood. Updating a string in any way
  creates a new `String` and assigns it back into the given reference; no other copies of the
  original `String` are affected.

- All expressions are evaluated left to right, with the exception any type of assignment statement.
  - In all assignment-like statements, the value (RHS) is evaluated before the reference (LHS)
  - This includes Set, Input, Read, Open, and other library functions that effectively perform an assignment.
  - This behavior is left unspecified.

- `Display` atatements only accept primitive types, not arrays or classes.
  - This behavior is left unspecified.
  - This could be supported in the future with some default string conversion rules.

- `For` and `For Each` loop variables must be simple variable references, not field or array element references.
  - This is implied by the book but not explicitly stated.
  - Avoids unspecified potential re-evaluations of complex reference expressions.
  - Could be changed in the future with a temp reference variable.

- `For` loop stop expression is re-evaluated on every iteration.
  - This behavior is left unspecified.

- `For` loop step expression must be a numeric constant.
  - This is implied by the book but not explicitly stated.
  - Avoids complicated and ambiguous code generation where the test expression (`<=` vs `>=`)
    might need to change depending on the value of the step expression.

- `For Each` loop array expression is re-evaluated twice on every iteration; once for the array
  length bounds check, once to assign the loop variable.
  - This behavior is left unspecified.
  - Could be changed in the future with a temp array reference variable.

- New library functions to convert number to String in an expression context.
  - `String integerToString(Integer n)`
  - `String realToString(Real n)`
  - Opposite of `stringToInteger`, `stringToReal`
  - In the book, this conversion only happens when invoking a `Display` statement, but it
    seems like an obvious oversight for useful string processing and formatting.

- Nested `Module` and `Function` declarations are not currently supported, but could be.
  - This is implied by the book but not explicitly stated.

- Programs with a `Module main()` are still allowed to execute arbitrary statements in the global block.
  - Global block statements first execute in lexical order, then `main()` is called afterward.
  - The book is unclear on whether mixing global statements and `Module main()` is legal.

- We implemented `Input` statements to loop until the user inputs a line that can be correctly
  parsed to the type of the input variable; e.g. a non-numeric input will loop and retry
  if the input variable is an `Integer`.

- By contrast, `Read` statements in an `InputFile` where the record data is the wrong type
  exit the program immediately with an exception.

- Arrays are filled with zero values on initialization when there is no initializer. When too
  few initializer expressions are provided, the remainder of the array is filled with zero values.
  - This behavior is left unspecified.

- Array initializers are currently the _only_ construct that may be parsed across multiple lines.
  - There are explicit examples of this in the book.
  - All other statements must appear on a single line!
  - This restriction may be relaxed in the future for other kinds of comma-separated lists.

- Arrays are deep copied when passed by value.
  - This is implied by the book but not explicitly stated.

- Array variables cannot be reassigned.
  - This is implied by the book but not explicitly stated.

- Array elements which are themselves Arrays (ie, part of a multidimensional array) cannot be
  reassigned either.
  - This is implied by the book but not explicitly stated.
  - For example, if `Integer table[3][4], row[4]`, you cannot `Set table[0] = row`.
  - However, such elements _may_ be passed as arguments, either by value or reference.
  - For example, `Call printValues(table[3])` is legal.
 
- Arrays cannot be the return value of a Function
  - This is implied by the book but not explicitly stated.

- Use `Call` to call the external library string modules `insert` and `delete`.
  - The book omits the `Call` keyword, which makes the syntax incompatible with the rest of the book.

- `Print` just outputs to stderr; there is no printer support.

- When reading and writing records with multiple fields using file I/O, there is no internal
  distinction between field and record separation. Writing or Reading multiple values in a
  single statement is equivalent to using multiple Write or Read statements.
