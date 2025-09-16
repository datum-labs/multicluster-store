# syntax=docker/dockerfile:1
FROM golang:1.24 as build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags mcstore_cacher -o /out/mc-apiserver ./examples/apiserver

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/mc-apiserver /mc-apiserver
USER nonroot:nonroot
ENTRYPOINT ["/mc-apiserver"]


