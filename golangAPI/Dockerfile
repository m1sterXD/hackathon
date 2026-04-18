FROM golang:1.26.1-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o app ./main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .
COPY data_extended_2.csv .

EXPOSE 8080

CMD ["./app"]