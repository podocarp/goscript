# goscript

A **work in progress** meta circular evaluator for a simple subset of golang.
It is like the bastard child of golang and javascript.
I wrote it to support user defined scripts in an application that processed
streams of metric data.
Right now it's good for just that: users can add a goscript script in yaml to
define simple functions that mutate incoming data. Examples are things like a
moving average, linear interpolation, basically raw loops and arithmetic and
nothing fancy.

Some benefits:
- Supposed to be quick and easy to write, even JS devs can write it.
- Parsed with the native `go/parser` parser.
- You can inject values into the global scope of the script.
- You can parse a function, save it, and execute it with different arguments later.
- It has to be fast and lightweight, no need for fancy features.



## Examples

Think of this as a JS but you are writing it in golang.
This is valid in `goscript`:
```go
func() {
    Fib := func (n) {
        if n < 2 {
            return n
        }
        return Fib(n-1) + Fib(n-2)
    }
    return Fib(10)
}()
```
Notice how types are not needed, and recursion can be done without pre-defining
the function variable.

Type promotions are also in place in a style of C/C++, most notably you can do
add ints to floats with no explicit casting. All in the name of convenience.

## Usage

Create a machine and evaluate code:
```go
func main() {
	m := machine.NewMachine()

	stmt := "func(a) { return a }(1)"
	res, err := m.ParseAndEval(stmt)
}
```
Because of the parser used you need to encapsulate all your code in a function
if you are going to write more than one line.

You can also make the function call at a later time:
```go
func main() {
	m := machine.NewMachine()

	stmt := "func(a) { return a }"
	res, err := m.ParseAndEval(stmt)

	resLit := res.Value.(*ast.FuncLit)
	callRes, err := m.CallFunction(res, []ast.Expr{
		machine.Number(1).ToLiteral(),
	})
}
```

Right now supplying and obtaining values are not too convenient, I'm working on
some wrappers for this.

## Comparisons with other solutions

There are many ways to implement user scripts in an application.

### Lua

One popular method is Lua scripts. Launching a Lua VM is way too heavy for my
use case. I only needed really simple computations, mostly single passes through
an array.

TODO: benchmark this

### Golang interpreters

There are many golang interpreters. One of the more developed ones are [yaegi](https://github.com/traefik/yaegi).
It is still slower. Following is a test of recursively finding the 30-th
Fibonnaci number (exponential runtime):
```
goos: linux
goarch: amd64
pkg: github.com/podocarp/goscript/benchmarks
cpu: AMD Ryzen 7 5800X 8-Core Processor
BenchmarkFib-16                1        3571864657 ns/op
BenchmarkYaegi-16              1        4627993028 ns/op
PASS
ok      github.com/podocarp/goscript/benchmarks 8.208s
```
Only one iteration is run, which eliminates any speed difference
caused by type checking and initialization (`goscript` has neither
but yaegi needs both).

I am not sure how yaegi works, but it has many bells and whistles that I don't
need. It is a good alternative if you want to actually run full go within go,
especially with things like the stdlib.

TODO: test the speed for injecting an array into the context and operating on
it.

## Builtins

Only a few builtin functions are available.

Those that work similar to golang:
- append
- len

Those with differences:
- make – currently does not accept any capacity arguments. This means you can
  only write `make([]float64)`.


## Missing features

This implementation of golang is incorrect in certain small areas but major
functionality should be correct as enforced in the tests.

Missing features from actual golang:
- Only very basic runtime type checking
- uint support is basically non existent, just use an int
- No multi return and multi assign
- No channels and goroutines
- No packages and imports
- You need to wrap scripts in a function if you have more than one line of code
  because of the parser.
- Builtins are not complete and sometimes differ from the those in go.

TODO:

- Maps
- Types
- Proper slices with sizes
