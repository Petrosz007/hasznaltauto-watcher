FROM golang:1.22-alpine3.19 as builder

WORKDIR /src

RUN apk add --update build-base

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=1 go build -o ./hasznaltauto-watcher


FROM alpine:3.19

WORKDIR /app

COPY --from=builder /src/hasznaltauto-watcher /app/hasznaltauto-watcher

WORKDIR /data

CMD ["/app/hasznaltauto-watcher"]

