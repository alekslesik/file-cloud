# Include variables from the .envrc file
include .envrc

#=====================================#
# HELPERS #
#=====================================#

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

#=====================================#
# DEVELOPMENT #
#=====================================#

## run: go run the cmd/* application
.PHONY: run
run:
	systemctl stop file-cloud
	go run ./cmd/file-cloud --env=development --port=443

## execute: execute the bin/ binary file
.PHONY: execute
execute: build
	systemctl stop file-cloud
	./bin/./file-cloud


#=====================================#
# UNIT SERVISE #
#=====================================#

## start: start the file-cloud.servise
.PHONY: unit.start
unit.start:
	systemctl start file-cloud

## stop: stop the file-cloud.servise
.PHONY: unit.stop
unit.stop:
	systemctl stop file-cloud

## restart: restart the file-cloud.servise
.PHONY: unit.restart
unit.restart:
	systemctl restart file-cloud

## status: status the file-cloud.servise
.PHONY: unit.status
unit.status:
	systemctl status file-cloud

#=====================================#
# DATABASE #
#=====================================#

## mysql: connect to the database using mysql
.PHONY: mysql
mysql:
	mysql -u web 'file_cloud' -p

## migrations.new name=$1: create a new database migration
.PHONY: migrations.new
migrations.new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=migrations ${name}

## migrations.up: apply all up database migrations
.PHONY: migrations.up
migrations.up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${WEB_DB_DSN} up

## migrations.down: apply all up database migrations
.PHONY: migrations.down
migrations.down:
	@echo 'Running down migrations...'
	migrate -path ./migrations -database ${WEB_DB_DSN} down

## migrations.force v=$1: do force migrations
.PHONY: migrations.force
migrations.force: confirm
	@echo 'Running force migrations v ${v}.'
	migrate -path ./migrations -database ${WEB_DB_DSN} force ${v}


#=====================================#
# QUALITY CONTROL #
#=====================================#

## audit: tidy dependencies and format, vet and test all code

## go fmt ./... : command to format all .go files in the project directory, according to the Go standard.
## go vet ./... : runs a variety of analyzers which carry out static analysis of your code and warn you
## go test -race -vet=off ./... : command to run all tests in the project directory
## staticcheck tool : to carry out some additional static analysis checks.
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	# staticcheck ./...
	# @echo 'Running tests...'
	# go test -race -vet=off ./...

## go mod tidy : prune any unused dependencies from the go.mod and go.sum files, and add any missing dependencies
## go mod verify : check that the dependencies on your computer (located in your module cache located at $GOPATH/pkg/mod)
## haven’t been changed since they were downloaded and that they match the cryptographic hashes in your go.sum file
## go mod vendor: copy the necessary source code from your module cache into a new vendor directory in your project root
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	# @echo 'Vendoring dependencies...'
	go mod vendor

#=====================================#
# BUILD #
#=====================================#

current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/api: build the cmd/api application
.PHONY: build
build: audit
	systemctl stop file-cloud
	go build -ldflags=${linker_flags} -o=./bin/file-cloud ./cmd/file-cloud
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=/var/www/file-cloud ./cmd/file-cloud
	cp -a -R tls /var/www/
	cp -a -R website /var/www/
	systemctl restart file-cloud
	# tail -f /var/www/logs/log.log

