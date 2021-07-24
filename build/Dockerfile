FROM golang:1.16.6-alpine3.14 as builder
RUN  \
     cd / && \
     apk update && \
     apk add --no-cache git ca-certificates make tzdata
WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd cmd
COPY internal internal
RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -v -a -o /alert-bot/alert-bot /build/cmd/alert-bot

FROM alpine:3.14
COPY --from=builder /alert-bot/alert-bot /
RUN apk add --no-cache ca-certificates tzdata tini
USER nobody
COPY templates /templates
EXPOSE 9087
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/alert-bot"]
