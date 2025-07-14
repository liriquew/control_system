FROM golang:1.24 AS build-stage

WORKDIR /app

COPY tasks_service/go.mod tasks_service/go.sum ./tasks_service/

# required for build (../ replace in go.mod)
COPY service_protos ./service_protos/

RUN cd tasks_service && go mod download && cd ..

COPY tasks_service ./tasks_service/

RUN cd tasks_service && CGO_ENABLED=0 GOOS=linux go build -o /tasks_service cmd/main.go

# Run Stage
FROM alpine:3.19 AS run-stage

WORKDIR /

COPY --from=build-stage /tasks_service /tasks_service
COPY --from=build-stage /app/tasks_service/config/docker_config.yaml /config/docker_config.yaml

EXPOSE 8080

ENTRYPOINT ["/tasks_service"]
