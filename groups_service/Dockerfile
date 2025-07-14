FROM golang:1.24 AS build-stage

WORKDIR /app

COPY groups_service/go.mod groups_service/go.sum ./groups_service/

# required for build (../ replace in go.mod)
COPY service_protos ./service_protos/

RUN cd groups_service && go mod download && cd ..

COPY groups_service ./groups_service/

RUN cd groups_service && CGO_ENABLED=0 GOOS=linux go build -o /groups_service cmd/main.go

# Run Stage
FROM gcr.io/distroless/base-debian11 AS run-stage

WORKDIR /

COPY --from=build-stage /groups_service /groups_service
COPY --from=build-stage /app/groups_service/config/docker_config.yaml /config/docker_config.yaml

EXPOSE 8080

ENTRYPOINT ["/groups_service"]
