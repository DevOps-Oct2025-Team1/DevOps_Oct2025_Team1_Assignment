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

RUN go build -o api-gateway

RUN chmod +x ./api-gateway
=======
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-gateway

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/api-gateway .
>>>>>>> dev

EXPOSE 8000

CMD ["./api-gateway"]

