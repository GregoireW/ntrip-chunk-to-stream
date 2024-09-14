FROM golang:1.23 AS builder

WORKDIR /go/src
COPY main.go .
RUN CGO_ENABLED=0 go build -o main main.go

FROM scratch

WORKDIR /

COPY --from=builder /go/src/main /main

ENTRYPOINT ["/main"]
CMD ["2101", "caster.centipede.fr:2101"]