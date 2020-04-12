#!/bin/bash
# error if unbound variable used
set -u

if [ -z "$SPEAKER_ADDRESS" ]; then
  errecho "bluetooth speaker address not set!"
  exit 1
fi

export SPEAKER_ADDRESS
connected=$(connected.sh)

if [ "$connected" = "no" ]; then echo 'not connected!' && exit 0; fi

echo "disconnect $SPEAKER_ADDRESS" | bluetoothctl &>/dev/null
