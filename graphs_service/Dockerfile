FROM golang:1.24 AS build-stage

WORKDIR /app

COPY graphs_service/go.mod graphs_service/go.sum ./graphs_service/

# required for build (../ replace in go.mod)
COPY service_protos ./service_protos/

RUN cd graphs_service && go mod download && cd ..

COPY graphs_service ./graphs_service/

RUN cd graphs_service && CGO_ENABLED=0 GOOS=linux go build -o /graphs_service cmd/main.go

# Run Stage
FROM alpine:3.19 AS run-stage

WORKDIR /

COPY --from=build-stage /graphs_service /graphs_service
COPY --from=build-stage /app/graphs_service/config/docker_config.yaml /config/docker_config.yaml

EXPOSE 8080

ENTRYPOINT ["/graphs_service"]
