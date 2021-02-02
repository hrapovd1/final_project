FROM scratch AS sys-mon

WORKDIR /sys-mon

COPY ./cmd/sys-mon/sys-mon /sys-mon/sys-mon

CMD ["/sys-mon/sys-mon"]
#--------------{ builder }--------------------
FROM golang:1.15 AS builder

RUN useradd -m -s /bin/bash go

WORKDIR /home/go
