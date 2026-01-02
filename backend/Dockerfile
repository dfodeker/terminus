# syntax=docker/dockerfile:1
ARG GO_VERSION=1.24.1

FROM golang:1.24.1-alpine AS build
WORKDIR /src

COPY backend/go.mod backend/go.sum ./backend/

RUN --mount=type=cache,target=/go/pkg/mod \
    cd backend && go mod download

# 2) Copy the rest of the repo
COPY . .
RUN echo "TARGETOS=$TARGETOS TARGETARCH=$TARGETARCH TARGETPLATFORM=$TARGETPLATFORM"
# 3) Build binaries
RUN --mount=type=cache,target=/go/pkg/mod \
    cd backend && go build -trimpath -o /out/api ./

RUN --mount=type=cache,target=/go/pkg/mod \
    cd backend && go build -trimpath -o /out/worker ./cmd/worker

RUN chmod +x /out/api /out/worker

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

COPY --from=build /out/api /api
COPY --from=build /out/worker /worker

EXPOSE 8080

# Build binaries for the target platform

