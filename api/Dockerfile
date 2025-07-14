FROM golang:1.24 AS build-stage

WORKDIR /app

COPY api/go.mod api/go.sum ./api/

# required for build (../ replace in go.mod)
COPY service_protos ./service_protos/

RUN cd api && go mod download && cd ..

COPY api ./api/

RUN cd api && CGO_ENABLED=0 GOOS=linux go build -o /api cmd/main.go

# Run Stage
FROM alpine:3.19 AS run-stage

WORKDIR /

COPY --from=build-stage /api /api
COPY --from=build-stage /app/api/config/docker_config.yaml /config/docker_config.yaml

EXPOSE 8080

ENTRYPOINT ["/api"]
