#!/bin/bash

errecho () { >&2 echo $@; }

if [ -z "$SPEAKER_ADDRESS" ]; then
  errecho "bluetooth speaker address not set!"
  exit 1
fi

export SPEAKER_ADDRESS
connected=$(connected.sh)

if [ "$connected" = "yes" ]; then exit 0; fi

CONN_MAX_TRY=${CONNECT_TRIALS:-1}

trial=1
while true; do
  echo "connect $SPEAKER_ADDRESS" | bluetoothctl &>/dev/null

  sleep 4

  CONNECTED=$(echo "info $SPEAKER_ADDRESS" | bluetoothctl 2>&1 | \
    grep -e Connected | awk '{ print $2 }')

  if [ "$CONNECTED" == "yes" ]; then
    exit 0
  else
    if [ $trial -eq $CONN_MAX_TRY ]; then
      errecho Failed to connect... Push scan button on speaker?
      exit 1
    fi
  fi
  let "trial += 1"
  sleep 1
done
