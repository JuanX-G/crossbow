# Contributing
To contribute fork the repository, make your changes and create a PR with the changes.
In your PR's include benchmarks, preferably attached to the message *and* as a file, following the format 
'benchmarks-PR_$(PR-title)_$(current-version).txt'. Populate the file by redirecting the output of the benchamrk command.
Make sure to include memory benchmarks.

In your PRs explain the change, make the explanation easy to understand; outline why it's useful.

## Useful Dev Commands

Runs all the tests. Use `-race` for race detection, but beware it may extend the test execution time drastically.
```Bash
go test -v ./...
```

Builds all the tests into a binary and includes goleak for leaked goroutines detection.
```Bash
go test -tags=leaktest -c -o tests
```

Compiles the tests bianry with goleak and runs each test separately outputing the result.
For easier execution you can use the helper script [leaktest](./tools/leaktest.sh)
```Bash
go test -tags=leaktest -c -o tests

for test in $(go test -list . | grep -E "^(Test|Example)"); do ./tests -test.run "^$test\$" &>/dev/null && echo -e "$test passed." || echo -e "\n! $test failed"; done
```

Runs all benchmarks. Use `-benchmem` to benchmark the memory usage and allocations made
```Bash
go test -bench=.
```

Runs all code generation. Use before any commits. Make sure to rerun after modyfying any enums that have '//go:generate (...)' above their type declaration.
```Bash
go generate .
```
