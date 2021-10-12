#!/bin/bash

fmt="%m/%d %H:%M:%S"
date "+Local: $fmt"
echo "---"
date -u "+UTC: $fmt"
TZ=Europe/London date "+London: $fmt"
TZ=America/Los_Angeles date -u "+LA: $fmt"
