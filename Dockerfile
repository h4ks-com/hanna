FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY main.go .
RUN go mod init ircbot && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ircbot main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/ircbot .
CMD ["./ircbot"]