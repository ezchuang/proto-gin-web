# Multi-stage build for Go Gin API

FROM golang:1.25 AS builder
WORKDIR /src

# Enable module downloads first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build static binary
ENV CGO_ENABLED=0
RUN go build -o /out/server ./cmd/api


FROM gcr.io/distroless/static:nonroot AS runner
WORKDIR /app

# Runtime env (overridable)
ENV PORT=8080
ENV APP_ENV=production

COPY --from=builder /out/server /app/server
COPY web /app/web
COPY internal/platform/http/templates /app/internal/platform/http/templates

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/server"]

