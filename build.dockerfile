FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS
ARG TARGETARCH
ARG CGO_ENABLED=1
ARG APP=todo
ARG PKG=./main.go
ARG VERSION=dev
ARG COMMIT=unknown

WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    EXT=$([ "$TARGETOS" = "windows" ] && echo ".exe" || echo "") && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=$CGO_ENABLED \
    go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/${APP}${EXT} ${PKG}

# Artifact stage so we can export binaries with buildx --output
FROM scratch AS artifact
ARG APP
COPY --from=build /out /out
