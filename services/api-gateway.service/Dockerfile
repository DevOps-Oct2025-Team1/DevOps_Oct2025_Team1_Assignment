FROM golang:1.25.5-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o api-gateway

RUN chmod +x ./api-gateway

EXPOSE 8000

CMD ["./api-gateway"]

