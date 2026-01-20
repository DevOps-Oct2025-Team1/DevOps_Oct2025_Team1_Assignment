FROM golang:1.25.5-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o auth-service

RUN chmod +x ./auth-service

EXPOSE 8001

CMD ["./auth-service"]

