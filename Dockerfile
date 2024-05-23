# Build stage
FROM golang:1.18 as builder

# Set environment variables for cross-compilation
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /go/src
COPY . .

# Build the Go application
RUN go build -a -o freeswitch_exporter

# Run stage
FROM scratch

COPY --from=builder /go/src/freeswitch_exporter /freeswitch_exporter

LABEL author="ZhangLianjun <z0413j@outlook.com>"

EXPOSE 9282
ENTRYPOINT ["/freeswitch_exporter"]
