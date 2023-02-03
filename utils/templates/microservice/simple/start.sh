#!/bin/bash

export DB_PORT=5432
export DB_USER=--dbuser--
export DB_PASSWORD=--dbpass--
export DB_NAME=--dbname--

export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_CACHE_ENABLED=no


if [ "$SPARTA_REDIS" == "y" ]
then
    REDIS_CACHE_ENABLED=yes
fi

PORT=8102
FOLDER=/home/gepisolo/go/projects/--module--

CMD=${FOLDER}/--module--

go build -o ${CMD} main.go

chmod a+x ${CMD}

 
${CMD} gateway --port $PORT


