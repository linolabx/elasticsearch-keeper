FROM golang:1.22.8-alpine AS builder

WORKDIR /app

COPY go.* ./
RUN go mod download
RUN go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o es-keeper

FROM alpine:3

RUN groupadd -g 1000 elasticsearch && \
    adduser --uid 1000 --gid 1000 --home /usr/share/elasticsearch elasticsearch && \
    adduser elasticsearch root && \
    chown -R 0:0 /usr/share/elasticsearch

WORKDIR /usr/share/elasticsearch

COPY --from=builder /app/es-keeper /usr/local/bin/es-keeper

ENTRYPOINT ["/usr/local/bin/es-keeper", "serve"]

USER 1000:0
