<<<<<<< HEAD
FROM golang:1.25.5-alpine
=======
FROM golang:1.25.5-alpine AS builder
>>>>>>> dev

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
<<<<<<< HEAD

RUN go build -o auth-service

RUN chmod +x ./auth-service
=======
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-service

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/auth-service .
>>>>>>> dev

EXPOSE 8001

CMD ["./auth-service"]

