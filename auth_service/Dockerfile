FROM golang:1.24 AS build-stage

WORKDIR /app

COPY auth_service/go.mod auth_service/go.sum ./auth_service/

# required for build (../ replace in go.mod)
COPY service_protos ./service_protos/

RUN cd auth_service && go mod download

COPY auth_service ./auth_service/

RUN cd auth_service && CGO_ENABLED=0 GOOS=linux go build -o /auth_service cmd/main.go

# Run Stage
FROM alpine:3.19 AS run-stage

WORKDIR /

COPY --from=build-stage /auth_service /auth_service
COPY --from=build-stage /app/auth_service/config/docker_config.yaml /config/docker_config.yaml

EXPOSE 8080

ENTRYPOINT ["/auth_service"]
