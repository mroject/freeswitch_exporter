# build
FROM golang:1.22 as builder
ARG LDFLAGS

WORKDIR /workspace
COPY go.mod go.sum /workspace/
RUN go mod download
COPY collector.go main.go prober.go /workspace/

RUN CGO_ENABLED=0 go build -a -ldflags "${LDFLAGS}" -o freeswitch_exporter && ./freeswitch_exporter --version

# run
FROM scratch

COPY --from=builder /workspace/freeswitch_exporter /freeswitch_exporter

LABEL author="ZhangLianjun <z0413j@outlook.com>"

EXPOSE 9282
ENTRYPOINT [ "/freeswitch_exporter" ]