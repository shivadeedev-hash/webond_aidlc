# syntax=docker/dockerfile:1

# ---- build stage ----
# Always runs on the build machine's own architecture and CROSS-compiles for the
# target (Go does this natively), so building the arm64 image on an x86 CI runner
# needs no emulation.
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum* ./
RUN go mod download

# Static, stripped, reproducible binary. TARGETARCH is set by `docker build
# --platform linux/arm64|amd64` (Fargate runs either); empty = the host's arch.
COPY . .
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/webond ./cmd/webond

# ---- runtime stage ----
# distroless/static: no shell, no package manager, non-root user.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/webond /app/webond
# Amazon RDS CA bundle (public certificates) so the API can verify the database
# server certificate: set DB_TLS_CA_FILE=/etc/ssl/rds/global-bundle.pem.
# Source: https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem
COPY certs/rds-global-bundle.pem /etc/ssl/rds/global-bundle.pem

EXPOSE 8080
USER nonroot:nonroot
# One-off commands reuse this image: `webond migrate`, `webond seed`.
ENTRYPOINT ["/app/webond"]
