#!/usr/bin/env bash

# migrate service is waiting for postgres start, then do migrate
# postgres-fill is waiting for migrate then load sql dump and write postgres-fill_ready file

# Important: concatenate shell command to block execution if one comand fail!!

# wait for postgres-fill_ready file in shared volume than write init_ready, this is the signal for test service to start testing
while [ ! -f /sharedContTest/postgres-fill_ready ]; do sleep 1; done && \
touch /sharedContTest/init_ready && \
tail -f /dev/null;
