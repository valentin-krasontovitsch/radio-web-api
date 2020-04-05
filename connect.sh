#!/bin/bash
# error if unbound variable used
set -u

errecho () { >&2 echo $@; }

if [ -z "$SPEAKER_ADDRESS" ]; then
  errecho "bluetooth speaker address not set!"
  exit 1
fi

export SPEAKER_ADDRESS
connected.sh &>/dev/null

if [ "$?" -eq "0" ]; then echo 'Already connected!' && exit 0;
else errecho not connected; echo; fi

CONN_MAX_TRY=${CONNECT_TRIALS:-1}

errecho "Will try to connext $CONN_MAX_TRY time(s)..."
errecho

trial=1
while true; do
  errecho Attempting to connect ...
  echo "connect $SPEAKER_ADDRESS" | bluetoothctl 2>&1 | xargs -L 1 errecho

  CONNECTED=$(echo "info $SPEAKER_ADDRESS" | bluetoothctl 2>&1 | xargs -L 1 echo \
    | grep -e Connected: | awk '{ print $2 }')

  if [ "$CONNECTED" == "yes" ]; then
    errecho We are connected!
    break
  else
    errecho Trial \# $trial
    if [ $trial -eq $CONN_MAX_TRY ]; then
      errecho Reached limit of $CONN_MAX_TRY failed connection trials
      errecho Failed to connect... Push scan button on speaker?
      exit 1
    fi
  fi
  let "trial += 1"
  sleep 1
done
