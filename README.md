# tcp_importer
Send Prometheus metrics to any TCP socket.  
The only format currently supported is JSON.  
The exporter will send all metric objects in a JSON Array.


## Installation

Prerequisites:

* [Go compiler](https://golang.org/dl/)

Installing:

    go get github.com/sighmir/tcp_importer
    go install github.com/sighmir/tcp_importer
    tcp_importer -c ./config.yml
