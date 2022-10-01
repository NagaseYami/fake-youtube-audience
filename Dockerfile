FROM golang:1.19.1-alpine as builder

WORKDIR /app
COPY . .
RUN go build main.go

FROM selenium/standalone-chrome:latest

COPY --from=builder /app/main /app/main

WORKDIR /app
ENTRYPOINT ["./main"]