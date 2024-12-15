FROM golang:1.23 AS build-go
WORKDIR /app
COPY api/ /app/
RUN go mod tidy && go build -o /app/main ./cmd/main.go

FROM python:3.12 AS build-python
WORKDIR /app
COPY predictions_service/ /app/
RUN pip install --no-cache-dir -r requirements.txt

FROM docker/compose:latest
COPY --from=build-go /app /app/api
COPY --from=build-python /app /app/predictions_service
COPY docker-compose.yaml /app/docker-compose.yaml
COPY migrations /app/migrations
COPY proto /app/proto
WORKDIR /app
CMD ["docker-compose", "up"]
