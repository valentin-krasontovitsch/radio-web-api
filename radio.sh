#!/bin/bash
# exit on non zero return
set -e
# error if unbound variable used
set -u

# TODO clean up output - send to /dev/null what we don't want to see
# and echo what we would like to see
export SPEAKER=${T5:-40:EF:4C:1D:37:F0}

export trials=5

/usr/local/bin/connect.sh

device_check="$(amixer -D bluealsa 2>/dev/null)"

if [ -z "${device_check}" ]; then
  echo Device bluealsa not available
  exit 1
fi

volume=${VOLUME:-40}

amixer -D bluealsa sset 'Audio Pro T5 - A2DP' \
  $volume 2>&1 >/dev/null

max_play_trials=5
trial=1
while true; do
  mplayer -ao alsa:device=bluealsa $MUSIC_SOURCE 2>&1 >/dev/null || \
    echo Trial \# $trial: Failed to play!
  if [ $trial -eq $max_play_trials ]; then
    echo Reached limit of $max_play_trials failed play trials
    exit 0
  fi
  # we don't want to keep retrying if we get killed because of a one hour timer
  if [ $SECONDS -gt 3500 ]; then
    echo time is up - not going to retry playing
    exit 0
  fi
  let "trial += 1"
  sleep 10
done
