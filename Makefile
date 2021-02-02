all: buildsrv buildcli

buildsrv: 
  @ if -z ${`docker images -q go:1.15`} ; then \
    docker build -f Dockerfile --tag go:1.15 --target builder .; \
  fi
  docker run --rm -v $(pwd):/home/go/project \
             -w /home/go --user 1000:1000 \
             --env GOPATH=/home/go --env GOCACHE=/home/go/gocache golang:1.15 \
             go build -race -ldflags "-s -w" -o bin/sys-mon cmd/sys-mon/main.go
            
buildcli:
  @ if -z ${`docker images -q go:1.15`} ; then \
    docker build -f Dockerfile --tag go:1.15 --target builder .; \
  fi
  docker run --rm -v $(pwd):/home/go/project \
             -w /home/go --user 1000:1000 \
             --env GOPATH=/home/go --env GOCACHE=/home/go/gocache golang:1.15 \
             go build -race -ldflags "-s -w" -o bin/sys-mon-cli cmd/sys-mon-cli/main.go