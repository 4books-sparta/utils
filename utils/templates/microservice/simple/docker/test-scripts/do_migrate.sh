#!/usr/bin/env bash

echo '----------------------------- Give POSTGRES 5 seconds to start -------------------------' && \
sleep 5
bash -c ". wait-for-it.sh --timeout=0 postgres:5432" && \
echo '----------------------------- POSTGRES MIGRATION START -------------------------' && \
migrate -source file://migrations -database postgres://--dbuser--:--dbpasss--@postgres:5432/--dbname--_test?sslmode=disable up && \
echo '----------------------------- POSTGRES MIGRATION END -------------------------' && \
sleep 1 && \
touch /sharedContTest/migrate_ready && \
tail -f /dev/null;