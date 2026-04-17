# ---------- build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# сначала зависимости
COPY go.mod go.sum ./
RUN go mod download

# потом код
COPY . .

# собираем бинарь
RUN go build -o app ./main

# ---------- runtime stage ----------
FROM alpine:latest

WORKDIR /app

# копируем бинарь
COPY --from=builder /app/app .

# копируем CSV (если нужен)
COPY data.csv .

EXPOSE 8080

CMD ["./app"]