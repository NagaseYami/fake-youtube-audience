FROM golang:1.19.1-alpine as builder

WORKDIR /app
COPY . .
ENV CGO_ENABLED 0
RUN go build -O fake-youtube-audience main.go

FROM selenium/standalone-chrome:latest

COPY --from=builder /app/fake-youtube-audience /app/fake-youtube-audience

WORKDIR /app
ENTRYPOINT ["./fake-youtube-audience"]