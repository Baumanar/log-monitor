ARG GO_VERSION="1.14.1"

# Builder image
FROM golang:${GO_VERSION}

# Sets workdir
WORKDIR /go/src/log-monitor
ADD . /go/src/log-monitor

# Create the default log file
RUN touch /tmp/access.log
RUN touch /dev/tty
# Installs dependencies
RUN go get -d -v ./...

# Compiles go app
RUN go build -o /go/bin/log-monitor



ENTRYPOINT ["/go/bin/log-monitor"]