ARG WITH_WEB=off

FROM node:26-alpine AS web-on

WORKDIR /web

RUN npm i -g pnpm@11.20.0

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./

RUN --mount=type=cache,target=/pnpm-store \
    pnpm config set store-dir /pnpm-store && \
    pnpm install --frozen-lockfile

COPY web/ ./

RUN pnpm build

FROM busybox AS web-off

RUN mkdir -p /web/.output/public

FROM web-${WITH_WEB} AS web

FROM golang:1.26-alpine AS build

WORKDIR /src

COPY api/go.mod api/go.sum ./
COPY api/cmd/ cmd/
COPY api/pkg/ pkg/
COPY --from=web /web/.output/public/ pkg/webui/dist/

ARG WITH_WEB
ARG TARGETOS=linux
ARG TARGETARCH=amd64

ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -mod=readonly -trimpath \
    -tags="$([ "${WITH_WEB}" = on ] && echo webui)" \
    -ldflags='-s -w' \
    -o /out/uichiyomiserver ./cmd/uichiyomiserver

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build /out/uichiyomiserver /uichiyomiserver

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT ["/uichiyomiserver"]
