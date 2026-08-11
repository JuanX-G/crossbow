# Contributing
To come

## Useful Dev Commands

Runs all the tests. Use `-race` for race detection, but beware it may extend the test execution time drastically.
```Bash
go test -v ./...
```

Builds all the tests into a binary.
```Bash
go test -c -o tests
```

Builds all the tests into a binary and includes goleak for leaked goroutines detection.
```Bash
go test -tags=leaktest -c -o tests
```

Compiles the tests bianry with goleak and runs each test separately outputing the result.
```Bash
go test -tags=leaktest -c -o tests

for test in $(go test -list . | grep -E "^(Test|Example)"); do ./tests -test.run "^$test\$" &>/dev/null && echo -e "\n$test passed." || echo -e "\n$test failed"; done
```

Runs all benchmarks
```Bash
go test -bench=.
```

Runs all code generation. Use before any commits. Make sure to rerun after modyfying any enums that have '//go:generate (...)' above their type declaration.
```Bash
go generate .
```
