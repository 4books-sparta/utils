#!/bin/bash

while [ ! -f /sharedContTest/migrate_ready ]; do sleep 1; done && \
# comment the line below for an empty db
psql -v 'ON_ERROR_STOP=on' postgres://--dbuser--:--dbpass-@postgres:5432/--dbname--_test < test_dump.sql && \
touch /sharedContTest/postgres-fill_ready && \
tail -f /dev/null;