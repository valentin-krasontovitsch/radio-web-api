#!/bin/bash
# error if unbound variable used
set -u

if [ -z "$SPEAKER_ADDRESS" ]; then
  echo "bluetooth speaker address not set!"
  exit 1
fi

export SPEAKER_ADDRESS
connected.sh

if [ "$?" -eq "0" ]; then echo 'Already connected!' && exit 0;
else echo not connected; echo; fi

CONN_MAX_TRY=${CONNECT_TRIALS:-1}

echo "Will try to connext $CONN_MAX_TRY time(s)..."
echo

trial=1
while true; do
  echo Attempting to connect ...
  echo "connect $SPEAKER_ADDRESS" | bluetoothctl 2>&1 | xargs -L 1 echo

  CONNECTED=$(echo "info $SPEAKER_ADDRESS" | bluetoothctl 2>&1 | xargs -L 1 echo \
    | grep -e Connected: | awk '{ print $2 }')

  if [ "$CONNECTED" == "yes" ]; then
    echo We are connected!
    break
  else
    echo Trial \# $trial
    if [ $trial -eq $CONN_MAX_TRY ]; then
      echo Reached limit of $CONN_MAX_TRY failed connection trials
      echo Failed to connect... Push scan button on speaker? >&2
      exit 1
    fi
  fi
  let "trial += 1"
  sleep 1
done
