# Only edit parser.go

## Create a recursive descent parser that can parse a shell language with the following syntax:
### Operators (symbol, internal name):
;, then
&&, and
||, or

### Special operators (symbol, internal name):
|, pipe
This operator should be unlike other operators in that the ast node will contain a list of calls with 2 calls corresponding to 1 pipe, 3 calls corresponding to 2 pipes etc.
eg.
callA "a" | callB "b" | callC "c"
is would be represented by a pipe ast node with 3 child call ast nodes.

### Calls:
Any sequence of 1 or more tokens that does not start with an operator.

### Parenthesis:
Used to group expressions.

### Examples:
(echo "hi" || echo "bye") && echo "success"
sleep 10; echo "awake now"

## Use the token type from tokenizer.go

## AST Nodes:
Should all implement an interface like:
```
type astNode interface {
    exec(pane *pane)
}
```
but the implementations must be empty for now.