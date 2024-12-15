FROM golang:1.23

WORKDIR /app

COPY . .

RUN go mod tidy

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

RUN go build ./cmd/main.go

ENV PORT=8080

EXPOSE 8080

CMD ["./main"]
