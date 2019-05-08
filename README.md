# tcp_exporter
Send Prometheus metrics to any TCP socket.  
The only format currently supported is JSON.  
The exporter will send all metric objects in a JSON Array.


## Installation

Prerequisites:

* [Go compiler](https://golang.org/dl/)

Installing:

    go get github.com/sighmir/tcp_exporter
    go install github.com/sighmir/tcp_exporter
    tcp_exporter -c ./config.yml
