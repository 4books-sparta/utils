#!/bin/ash

DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSL"

migrate -path /migrations  -database "$DATABASE_URL"  up