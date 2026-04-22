FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /webserver ./cmd/webserver

FROM alpine:latest
WORKDIR /app
COPY --from=builder /webserver ./webserver
COPY --from=builder /app/config ./config
COPY --from=builder /app/frontend ./frontend
CMD ["/app/webserver"]
