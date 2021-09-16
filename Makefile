ENV = prod staging local

SHELL=/bin/bash

ifeq (${ENV}, prod)
	include .prod.env
endif
export


check-int:
ifeq ($(ENV),int)
	@echo "LOCAL ENV active"
else
	$(error == env must be LOCAL DEV==)	
endif

# DEV

run-server:
	docker-compose -f dc-int.yml up -d db-sqlm && \
	go run ./cmd/main.go

# INT

test-int-local: check-int
	docker-compose -f dc-int.yml up -d db-sqlm && \
	go test -count=1 -v ./serverlib/integration -test.run=${args}

test-int-ga:
	go test -count=1 -v ./serverlib/integration -test.run=Test

# RELEASE

release:
	git tag -a v${tag} -m "release v${tag}" && git push --follow-tags

run-release-local:
	./dist/sql-manager-test_linux_amd64/sql-manager-auth