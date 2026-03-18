FROM golang:1.26 AS mod

WORKDIR /app

RUN --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download


FROM golang:1.26 AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
ARG BUILT_BY=docker

WORKDIR /app

ENV CGO_ENABLED=0

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
  go build -ldflags="-s -w -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE} -X main.builtBy=${BUILT_BY}" -trimpath -o entrypoint

FROM scratch AS final

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY LICENSE .
COPY --from=builder /app/entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
