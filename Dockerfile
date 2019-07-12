FROM golang:1.10.3-stretch AS builder

# add proxy
ENV https_proxy http://172.16.133.102:1087

RUN go get -u github.com/prometheus/client_golang/prometheus
RUN go get -u github.com/prometheus/log
RUN go get -u github.com/prometheus/client_golang/prometheus/promhttp

# unset proxy
RUN unset https_proxy
ENV https_proxy ""

COPY . /go/src/git.fccs.com/fccs_zj/ssl-exporter
WORKDIR /go/src/git.fccs.com/fccs_zj/ssl-exporter
RUN go build -o ssl-exporter

FROM debian:stretch
COPY --from=builder /go/src/git.fccs.com/fccs_zj/ssl-exporter/ssl-exporter /usr/bin/ssl-exporter
ENTRYPOINT ["/usr/bin/ssl-exporter"]

EXPOSE 8080
