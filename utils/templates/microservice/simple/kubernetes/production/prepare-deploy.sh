#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_DIR=${SCRIPT_DIR}/../..
ENV_DIR=${PROJECT_DIR}/docker/config/env/prod

source ${ENV_DIR}/db.env

BASE=${SCRIPT_DIR}/deployment.base
DEPLOYMENT=${SCRIPT_DIR}/deployment.yaml

cp ${BASE} ${DEPLOYMENT}

sed -i "s/--DB_HOST--/${DB_HOST}/g" ${DEPLOYMENT}
sed -i "s/--DB_PORT--/${DB_PORT}/g" ${DEPLOYMENT}
sed -i "s/--DB_NAME--/${DB_NAME}/g" ${DEPLOYMENT}
echo " - DB configured"

sed -i "s|--migration-image--|--aws-account-id--.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms---service-name---migrations:latest|g" ${DEPLOYMENT}
BASE_IMAGE=--aws-account-id--.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms---service-name--
IMAGE=${BASE_IMAGE}:build-$BITBUCKET_COMMIT
sed -i "s|--microservice-image--|${IMAGE}|g" ${DEPLOYMENT}

echo " - DEPLOYMENT config ready"

