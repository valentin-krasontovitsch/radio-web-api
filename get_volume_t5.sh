#!/bin/bash
set -e
CONTENTS=$(amixer -D bluealsa scontents)
if [ -z "$CONTENTS" ]; then
  echo Failed to acces bluealsa simple contents via amixer >&2
  echo Are you connected? >&2
  exit 1
fi
export CONTENTS
VOLUME=$(echo $CONTENTS | tail -n1 | sed -E 's/.*\[([0-9]+)%\].*/\1/')
echo $VOLUME
exit 0
