# live-exporter
A Prometheus exporter to be used with the TCP Input plugin from Intelie Live

## Building and running

Prerequisites:

* [Go compiler](https://golang.org/dl/)

Installing:

    go get github.com/sighmir/live-exporter
    go install github.com/sighmir/live-exporter
    live-exporter -c ./config.yml
