FROM --platform=$BUILDPLATFORM node:26-alpine AS web

WORKDIR /web

RUN npm i -g pnpm@11.20.0

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./

RUN --mount=type=cache,target=/pnpm-store \
    pnpm config set store-dir /pnpm-store && \
    pnpm install --frozen-lockfile

COPY web/ ./

RUN pnpm build

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

WORKDIR /src

COPY api/go.mod api/go.sum ./
COPY api/cmd/ cmd/
COPY api/pkg/ pkg/
COPY --from=web /web/.output/public/ pkg/webui/dist/

ARG TARGETOS=linux
ARG TARGETARCH=amd64

ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -mod=readonly -trimpath \
    -tags=webui \
    -ldflags='-s -w' \
    -o /out/uchiyomiserver ./cmd/uchiyomiserver

FROM gcr.io/distroless/static-debian13

COPY --from=build /out/uchiyomiserver /uchiyomiserver

EXPOSE 3000

ENTRYPOINT ["/uchiyomiserver"]
