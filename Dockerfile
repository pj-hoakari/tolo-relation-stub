# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags='-s -w' -o /out/relation-stub ./cmd/server


FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/relation-stub /usr/local/bin/relation-stub

ENV GIN_MODE=release \
    RELATION_ADDR=:8080

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/relation-stub"]
