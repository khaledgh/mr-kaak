# Multi-stage build. BINARY selects which entrypoint to compile (api|worker),
# so one Dockerfile produces both images via the docker-compose build args.

# ---- build stage ----
FROM golang:1.24-alpine AS build
ARG BINARY=api
WORKDIR /src

# Cache module downloads separately from source for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static build, stripped, version stamped from VCS if available.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/app ./cmd/${BINARY}

# ---- runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -u 10001 appuser
USER appuser
WORKDIR /app
COPY --from=build /out/app /app/app

EXPOSE 8080
ENTRYPOINT ["/app/app"]
