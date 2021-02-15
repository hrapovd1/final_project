#--------------{ builder }--------------------
FROM golang:1.15 AS builder
ENV GOPATH='/go'

COPY ./cmd /go/project/cmd
COPY ./pkg /go/project/pkg
COPY ./go.* /go/project/

RUN cd /go/project && go build -o ./bin/sys-mon ./cmd/sys-mon
#-------------------------------------------
FROM golang:1.15

WORKDIR /sys-mon

COPY --from=builder /go/project/bin/sys-mon /sys-mon/sys-mon

ENTRYPOINT ["/sys-mon/sys-mon"]
