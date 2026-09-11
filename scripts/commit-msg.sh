#!/bin/sh
head -1 "$1" | grep -Eq '^(Merge |Revert |(feat|fix|chore|docs|ci|refactor|test)(\([a-z0-9-]+\))?!?: .+)' && exit 0
echo "bad commit message"
echo "format: <type>(<scope>): <summary>   scope optional"
echo "types:  feat fix chore docs ci refactor test"
exit 1
