#!/bin/bash
# error if unbound variable used
set -u

errecho () { >&2 echo $@; }

if [ -z "$SPEAKER_ADDRESS" ]; then
  errecho "bluetooth speaker address not set!"
  exit 1
fi

export SPEAKER_ADDRESS
connected=$(connected.sh)

if [ "$connected" = "yes" ]; then echo 'Already connected!' && exit 0;
else errecho not connected; errecho; fi

CONN_MAX_TRY=${CONNECT_TRIALS:-1}

errecho "Will try to connext $CONN_MAX_TRY time(s)..."
errecho

trial=1
while true; do
  errecho Trial \# $trial
  errecho Attempting to connect ...
  echo "connect $SPEAKER_ADDRESS" | bluetoothctl &>/dev/null

  sleep 4

  CONNECTED=$(echo "info $SPEAKER_ADDRESS" | bluetoothctl 2>&1 | \
    grep -e Connected | awk '{ print $2 }')

  if [ "$CONNECTED" == "yes" ]; then
    errecho We got connected!
    break
  else
    if [ $trial -eq $CONN_MAX_TRY ]; then
      errecho Reached limit of $CONN_MAX_TRY failed connection trials
      errecho Failed to connect... Push scan button on speaker?
      exit 1
    fi
  fi
  let "trial += 1"
  sleep 1
done
