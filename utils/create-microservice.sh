#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NOCOLOR='\033[0m'
DONE=${GREEN}OK${NOCOLOR}

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
LAST_UTILS_TAG=$(git -c 'versionsort.suffix=-' ls-remote --exit-code --refs --sort='version:refname' --tags https://github.com/4books-sparta/utils '*.*.*' | tail --lines=1 | cut --delimiter='/' --fields=3)
# LAST_UTILS_TAG=v0.7.21

source .env

source ${STAGING_ENV_FILE}
STAGE_DB_USER=$DB_USER
STAGE_DB_PASS=$DB_PASSWORD
STAGE_DB_HOST=$DB_HOST



TEMPLATE=${SCRIPT_DIR}/templates/microservice/$MODULE_TYPE.tgz

DEST_FOLDER=${BASE_PATH}/${MODULE}

echo "Creating folder... ${DEST_FOLDER}"
mkdir -p $DEST_FOLDER

CHECK_FILE=${DEST_FOLDER}/main.go

echo "Verifying it's empty..."
if test -f "$CHECK_FILE"; then
    echo -e "Folder is ${RED}not empty${NOCOLOR}... exit"
    exit 1
fi
echo -e $DONE

echo "Moving template files"
tar -xzf ${TEMPLATE} -C $DEST_FOLDER

echo "Creating main service..."
mv ${DEST_FOLDER}/pkg/service/test-svc ${DEST_FOLDER}/pkg/service/${SERVICE_NAME}
echo -e $DONE

echo "Replacing placeholders..."
for f in $(find ${DEST_FOLDER}/ -type f); do
    if [[ "$f" == *".git/"* ]]; then
	continue
    fi
    if [[ "$f" == *".idea/"* ]]; then
	continue
    fi
    sed -i "s|--module--|${MODULE}|g" $f
    sed -i "s|--utils-tag--|${LAST_UTILS_TAG}|g" $f
    sed -i "s|--service-name--|${SERVICE_NAME}|g" $f
    sed -i "s|--prod-dbname--|${PRODDBNAME}|g" $f
    sed -i "s|--stage-dbname--|${STAGE_DB_HOST}|g" $f
    sed -i "s|--stage-dbuser--|${STAGE_DB_USER}|g" $f
    sed -i "s|--stage-dbpass--|${STAGE_DB_PASS}|g" $f
    sed -i "s|--prom-user--|${PROMUSER}|g" $f
    sed -i "s|--prom-pass--|${PROMPASS}|g" $f
    sed -i "s|--dbuser--|${DEVDBUSER}|g" $f
    sed -i "s|--dbpass--|${DEVDBPASS}|g" $f
    sed -i "s|--aws-account-id--|${AWS_ACCOUNT_ID}|g" $f
    sed -i "s|--dbname--|${DEVDBNAME}|g" $f
    sed -i "s|--port-name--|${SVC_PORT_NAME}|g" $f
done;
echo -e $DONE

echo "Preparing gateway..."
sed -i "s|--dbuser--|${DEVDBUSER}|g" ${DEST_FOLDER}/start.sh
sed -i "s|--dbpass--|${DEVDBPASS}|g" ${DEST_FOLDER}/start.sh
sed -i "s|--dbname--|${DEVDBNAME}|g" ${DEST_FOLDER}/start.sh
echo -e $DONE

echo "Cleaning module files..."
cd $DEST_FOLDER
go mod tidy
echo -e $DONE


echo "Please prepare the following 4 ECR folders:"
echo -e "${CYAN} XXXXXXXXXXXX.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms-${SERVICE_NAME}-migrations  ${NOCOLOR}"
echo -e "${CYAN} XXXXXXXXXXXX.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms-${SERVICE_NAME}  ${NOCOLOR}"
echo -e "${CYAN} XXXXXXXXXXXX.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms-${SERVICE_NAME}-migrations-staging  ${NOCOLOR}"
echo -e "${CYAN} XXXXXXXXXXXX.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms-${SERVICE_NAME}-staging  ${NOCOLOR}"
echo -e $DONE

echo "Please create 4 databases (dev, minikube, staging, prod):"
echo -e "${CYAN} CREATE DATABASE ${DEVDBNAME} OWNER ${DEVDBUSER}; ${NOCOLOR}"
echo -e "${CYAN} CREATE DATABASE ${PRODDBNAME} OWNER king; ${NOCOLOR}"

source $PROD_ENV_FILE

echo "Please create this ENV vars in bitbucket pipelines:"
echo -e "${CYAN} AWS_ACCESS_KEY_ID: ${AWS_DEPLOYER_ACCESS_KEY}  ${NOCOLOR}"
echo -e "${CYAN} KUBE_CONFIG_XXX: ${NOCOLOR}"
echo -e "${CYAN} AWS_DEFAULT_REGION:  ${NOCOLOR}"
echo -e "${CYAN} AWS_SECRET_ACCESS_KEY: ${AWS_DEPLOYER_SECRET_KEY} ${NOCOLOR}"


echo -e $DONE
