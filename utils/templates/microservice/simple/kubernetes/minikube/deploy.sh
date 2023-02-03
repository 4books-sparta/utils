#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_DIR=${SCRIPT_DIR}/../..

source ${SCRIPT_DIR}/.env

# Ensure context is minikube or exit !
CUR_CTX=$(kubectl config current-context)
if [ "$CUR_CTX" != "minikube" ]; then
    echo -e "Your context is \033[1mNOT minikube\033[0m! (\033[0;36m${CUR_CTX}\033[0m)"
    exit 1
fi

# Build migrations image
MIGRATIONS_DIR=${PROJECT_DIR}/docker/images/migrations
DATA_DIR=${MIGRATIONS_DIR}/data
cp -r ${PROJECT_DIR}/migrations/* ${DATA_DIR}

cd $MIGRATIONS_DIR

docker build -f Dockerfile -t ${MIGRATIONS_IMG_NAME}:latest .

rm -r ${DATA_DIR}/postgres/*.sql

# Load migrations image into minikube
minikube image load ${MIGRATIONS_IMG_NAME}:latest

# Build microservice image
cd $PROJECT_DIR

docker build -f docker/images/microservice/Dockerfile -t ${SERVICE_IMG_NAME}:latest .
# Load microservice image
minikube image load ${SERVICE_IMG_NAME}:latest

# Deploy
cd $SCRIPT_DIR
kubectl apply -f deployment.yaml