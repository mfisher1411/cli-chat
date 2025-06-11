FROM golang:1.24.3-alpine3.21 AS builder

COPY libraries /app/libraries

COPY services/chat_server /app/services/chat_server

WORKDIR /app/services/chat_server

RUN go mod download
RUN go build -o ./bin/main ./cmd/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/services/chat_server/bin/main ./main

CMD ["./main"]