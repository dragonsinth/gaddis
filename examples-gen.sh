#!/bin/bash

set -e

ANYERROR=0

for gadFile in $(find examples -name '*.gad'); do
  echo "== Running $gadFile"
  if go run ./cmd/gaddis -t "$gadFile"; then
    echo "== ok"
  else
    echo "== ERROR: $gadFile"
    ANYERROR=1
  fi
done

if [[ $ANYERROR -eq 1 ]]; then
  echo "There were errors"
  exit 1
fi
