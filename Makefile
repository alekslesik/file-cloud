# Include variables from the .envrc file
include .envrc

production_host_ip = "188.120.228.254"

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
# RUN #
#=====================================#

## run: go run the cmd/* application
.PHONY: run
run:
	direnv allow
	go run ./cmd/file-cloud -env=development -port=8080

## run.bin: execute the bin/ binary file
.PHONY: run.bin
run.bin: build
	direnv allow
	./bin/linux_amd64/file-cloud -port=8080

## run.service: go run local service
.PHONY: run.service
run.service: build
	sudo cp bin/file-cloud /var/www
	sudo cp -r tls/ /var/www/
	sudo cp -r tmp/ /var/www/
	sudo cp -r website /var/www/
	sudo cp -r .envrc /var/www/


#=====================================#
# UNIT SERVISE #
#=====================================#

## create: create the file-cloud.servise
.PHONY: unit.create
unit.create:
	sudo cp /home/kasian/go/src/githhub.com/alekslesik/file-cloud/remote/production/file-cloud.service /etc/systemd/system/
	sudo systemctl enable file-cloud
	sudo systemctl restart file-cloud
	sudo systemctl status file-cloud


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
	mysql -u 'file_cloud' -p

## migrations.new name=$1: create a new database migration
.PHONY: migrations.new
migrations.new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=migrations ${name}

## migrations.up: apply all up database migrations
.PHONY: migrations.up
migrations.up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${MIGRATE_DSN} up

## migrations.down: apply all up database migrations
.PHONY: migrations.down
migrations.down:
	@echo 'Running down migrations...'
	migrate -path ./migrations -database ${MIGRATE_DSN} down

## migrations.force v=$1: do force migrations
.PHONY: migrations.force
migrations.force: confirm
	@echo 'Running force migrations v ${v}.'
	migrate -path ./migrations -database ${MIGRATE_DSN} force ${v}


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
linker_flags = -s -X main.buildTime=${current_time} -X main.version=${git_description}

## build: build the cmd/api application
.PHONY: build
build: audit
	@echo 'Building cmd/file-cloud...'
	go build -ldflags "${linker_flags} -linkmode=external -extldflags '-static'" -o=./bin/file-cloud ./cmd/file-cloud

#=====================================#
# PRODUCTION #
#=====================================#

## production.connect: connect to the production server
.PHONY: production.connect
production.connect:
	ssh root@${production_host_ip}

## production.deploy: deploy the api to production
.PHONY: production.deploy
production.deploy: build
	rsync -rP --delete ./bin/file-cloud ./migrations ./tls ./website ./.envrc root@${production_host_ip}:~/web
	ssh -t root@${production_host_ip} 'migrate -path ~/web/migrations -database mysql://file_cloud:Todor1990@tcp/file_cloud up'