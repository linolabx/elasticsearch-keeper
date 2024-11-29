FROM golang:1.22.8-alpine AS builder

WORKDIR /app

COPY go.* ./
RUN go mod download
RUN go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o es-keeper

FROM alpine:3

COPY --from=builder /app/es-keeper /usr/local/bin/es-keeper

ENTRYPOINT ["/usr/local/bin/es-keeper", "serve"]