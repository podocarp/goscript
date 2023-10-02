# goscript

A meta circular evaluator for a simple subset of golang.
It is like the bastard child of golang and javascript.
I wrote it to support user defined scripts in an application that processed
streams of metric data.

Some benefits:
- Supposed to be quick to write, even JS devs can write it
- Parsed with the native `go/parser` parser, so you can be sure it is correct,
  since golang is literally implemented with it
- You can inject values into the global scope of the script. This is really
  important since it allows for scripts that actually do work.
- It has to be fast and lightweight.

## Comparisons

There are many ways to implement user scripts in an application.

### Lua

One popular method is Lua scripts. Launching a Lua VM is way too heavy for my
use case. I only needed really simple computations, mostly single passes through
an array.

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
It is so slow there is only 1 iteration, which eliminates any speed difference
caused by type checking and initialization (`goscript` has neither).

`goscript` is fast because it ignores like half the go spec.

TODO: test the speed for injecting an array into the context and operating on
it.


## Differences

This implementation of golang is incorrect in certain small areas but major
functionality should be correct as enforced in the tests.
Notable differences from actual golang:
- No types, they are actively ignored
- No multi return and multi assign
- All numbers and booleans are actually floats
- No packages and imports
- You need to wrap scripts in a function because the parser only works on
  expressions.

##

TODO:

- Maps
- Inject values
- Basic runtime typing
- Actual booleans
