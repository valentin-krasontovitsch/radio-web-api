#!/bin/bash

if [ -z "$SPEAKER_ADDRESS" ]; then
  errecho "bluetooth speaker address not set!"
  exit 1
fi

export SPEAKER_ADDRESS

connected=$(connected.sh)
if [ "$connected" = "no" ]; then echo 'not connected!' && exit 0; fi

echo "disconnect $SPEAKER_ADDRESS" | bluetoothctl &>/dev/null

trial=1
max_trials=5
while true; do
  connected=$(connected.sh)
  if [ "$connected" = "no" ]; then
    exit 0
  else
    if [ $trial -eq $max_trials ]; then
      errecho "Failed to disconnect?!"
      exit 1
    fi
  fi
  let "trial += 1"
  sleep 1
done
