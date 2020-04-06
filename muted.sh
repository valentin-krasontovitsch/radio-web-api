#!/bin/bash
set -e
CONTENTS=$(amixer -D bluealsa scontents)
if [ -z "$CONTENTS" ]; then
  echo Failed to acces bluealsa simple contents via amixer >&2
  echo Are you connected? >&2
  exit 1
fi
export CONTENTS
STATUS=$(echo $CONTENTS | tail -n1 | \
  sed -E 's/.*\[\([a-z]\+\)\]$/\1/')
if [ "$STATUS" = "off" ]; then
  echo true
else
  echo false
fi
exit 0
