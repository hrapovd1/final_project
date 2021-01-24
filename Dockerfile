FROM busybox:latest

WORKDIR /sys-mon

COPY ./cmd/sys-mon/sys-mon /sys-mon/sys-mon

CMD ["/sys-mon/sys-mon"]
