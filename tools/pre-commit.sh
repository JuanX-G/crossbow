#!/usr/bin/env bash

echo -e "# running the pre commit script"

set -euo pipefail

./tools/fmt.sh x

./tools/testall.sh c b

./tools/leaktest.sh c

echo -e "\n # pre-commit script done"