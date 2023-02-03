# go

.PHONY: build
build:
	go build main.go

.PHONY: up
up:
	env TZ=UTC \
	go run main.go service \
	 --port 8101

.PHONY: build1.19
build1.19:
	env TZ=UTC \
	go1.19 build main.go

.PHONY: up1.19
up1.19:
	env TZ=UTC \
	go1.19 run main.go service \
	--port 8101

.PHONY: clean
clean:
	go clean

.PHONY: deps
deps:
	dep ensure -v

# docker

TAG=$(shell date +%y%m%d%H%M)
BASE_IMAGE=--aws-account-id--.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms---service-name--
BASE_FILE=docker/Dockerfile
MIGRATION_IMAGE=--aws-account-id--.dkr.ecr.eu-west-1.amazonaws.com/sparta-ms---service-name---migrations
MIGRATION_FILE=docker/images/migrations/Dockerfile
DEV_DB_COMPOSE=docker/dev_db.yml

.PHONY: docker-mi-build
docker-mi-build:
	docker build -t ${MIGRATION_IMAGE} -f ${MIGRATION_FILE} .

.PHONY: docker-mi-push
docker-mi-push:
	docker push ${MIGRATION_IMAGE}

.PHONY: docker-db-up
docker-db-up:
	docker-compose -p 4books---service-name-- -f ${DEV_DB_COMPOSE} up

.PHONY: docker-db-down
docker-db-down:
	docker-compose -p 4books---service-name-- -f ${DEV_DB_COMPOSE} down

.PHONY: docker-db-in
docker-db-in:
	docker exec -it 4books---service-name--_db psql -U --dbuser-- -d --dbname--


MIGRATIONS_REL=migrations/postgres
MIGRATIONS_DIR=file://${MIGRATIONS_REL}

CONNECT_URL=postgres://--dbuser--:--dbpass--@localhost:5432/--dbname--?sslmode=disable
CONNECT_URL_DOCKER=postgres://--dbuser--:--dbpass--@localhost:6544/--dbname--?sslmode=disable

.PHONY: migrate
migrate:
	migrate -source ${MIGRATIONS_DIR} -database ${CONNECT_URL} up

.PHONY: migrate-docker
migrate-docker:
	migrate -source ${MIGRATIONS_DIR} -database ${CONNECT_URL_DOCKER} up

.PHONY: rollback
rollback:
	migrate -source ${MIGRATIONS_DIR} -database ${CONNECT_URL_DOCKER} down 1

.PHONY: rollbackall
rollbackall:
	migrate -source ${MIGRATIONS_DIR} -database ${CONNECT_URL} down

# Remove all postgres data:
# docker volume rm 4books_dbData

.PHONY: create-migration
create-migration:
	migrate create -ext sql -dir ${MIGRATIONS_REL} ${MIGRATION_NAME}

.PHONY: force-version
force-version:
	migrate -source ${MIGRATIONS_DIR} -database ${CONNECT_URL} force ${VERSION}

# test
DEV_TEST_COMPOSE=docker/dev_test.yml
DEV_TEST_COMPOSE_SINGLE=docker/dev_single_test.yml

.PHONY: test-dev
test-dev:
	docker-compose -f ${DEV_TEST_COMPOSE_SINGLE} up --build --abort-on-container-exit

.PHONY: test-up-no-exit
test-up-no-exit:
	DONT_EXIT_ON_END=yessa docker-compose -f ${DEV_TEST_COMPOSE} up --build --abort-on-container-exit

.PHONY: test-up
test-up:
	DONT_EXIT_ON_END=no docker-compose -f ${DEV_TEST_COMPOSE} up --build --abort-on-container-exit
	docker-compose -f ${DEV_TEST_COMPOSE} down -v

.PHONY: test-down
test-down:
	docker-compose -f ${DEV_TEST_COMPOSE} down -v

.PHONY: test-build
test-build:
	docker-compose -f ${DEV_TEST_COMPOSE} build

.PHONY: test-db-in
test-db-in:
	docker exec -it docker_postgres_1 psql -U --dbuser-- -d --dbname--_test