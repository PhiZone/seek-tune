#!/usr/bin/env bash

go build -tags netgo -ldflags '-s -w' -o app
./app serve -p 2338 3>&1 1>&2 2>&3 3>&-
