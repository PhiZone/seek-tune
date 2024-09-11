#!/usr/bin/env bash

HTTP_PID=$(sudo lsof -t -i:2338)

if [ -n "$HTTP_PID" ]; then
  sudo kill -9 $HTTP_PID
fi
