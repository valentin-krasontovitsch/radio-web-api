#!/bin/bash
echo info $SPEAKER_ADDRESS | bluetoothctl 2>/dev/null | grep Connected | \
  awk '{print $2}' | tr -d ' '
