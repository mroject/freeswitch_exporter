# build
FROM golang:1.20 as builder

WORKDIR /go/src
COPY . /go/src/
RUN CGO_ENABLED=0 go build -a -o freeswitch_exporter

# run
FROM scratch

COPY --from=builder /go/src/freeswitch_exporter /freeswitch_exporter

LABEL author="ZhangLianjun <z0413j@outlook.com>"

EXPOSE 9282
ENTRYPOINT [ "/freeswitch_exporter" ]
