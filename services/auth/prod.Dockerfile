FROM golang:1.24.3-alpine3.21 AS builder

COPY libraries /app/libraries

COPY services/auth /app/services/auth

WORKDIR /app/services/auth

RUN go mod download
RUN go build -o ./bin/main ./cmd/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/services/auth/bin/main ./main

CMD ["./main"]