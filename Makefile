.DEFAULT_GOAL := help
UID := $(shell id -u)
help:
	@echo "Сборка docker image сервера:\n\tmake docker\nСборка сервера и клиента в текущем окружении:\n\tmake all"

all: buildsrv buildcli

docker: 
	docker build -f Dockerfile --tag sys-mon:v0.2 .
            
buildsrv:
	docker run --rm -v $$(pwd):/go/project \
	           --env GOPATH=/go \
		   --workdir /go/project \
		   golang:1.15 \
        go build -o ./bin/sys-mon ./cmd/sys-mon/
	docker run --rm -v $$(pwd):/sys-mon golang:1.15 \
	           chown $(UID):$(UID) /sys-mon/bin/sys-mon

buildcli:
	docker run --rm -v $$(pwd):/go/project \
	           --env GOPATH=/go \
		   --workdir /go/project \
		   golang:1.15 \
        go build -o ./bin/sys-mon-cli ./cmd/sys-mon-cli/
	docker run --rm -v $$(pwd):/sys-mon golang:1.15 \
	           chown $(UID):$(UID) /sys-mon/bin/sys-mon-cli
