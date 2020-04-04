#!/bin/bash

PLAYER=mplayer

set -e
set -u

PID="$(pgrep $PLAYER)"

if [ -z "$PID" ]; then
  echo $PLAYER pid not found, is it running?
  exit 1
fi

kill -9 $PID

