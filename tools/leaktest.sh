#!/usr/bin/env bash 

echo -e "testing goroutine leaks\n"

clean=${1-default}
set -euo pipefail

if [ $clean = "c" ]; then 
    go clean -testcache
fi

go test -tags=leaktest -c -o tests # Leak test is the builtag for including the detector

failed=0
fail() {
    $failed=1
    local test = $1
    echo "[!!]: $test failed"
}

for test in $(go test -list . | grep -E "^(Test|Example)"); do ./tests -test.run "^$test\$" &>/dev/null && echo -e "$test passed." || (fail $test); done

if [ $failed -eq 1 ]; then 
    echo "FAIL"
fi