# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X github.com/42atomys/webhooked.Version=${VERSION} -X github.com/42atomys/webhooked.GitCommit=${COMMIT} -X github.com/42atomys/webhooked.BuildDate=${BUILD_DATE}" \
    -o webhooked \
    ./cmd/webhooked/webhooked.go

# Final stage
FROM scratch

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /build/webhooked /webhooked

# Create non-root user
COPY --from=builder /etc/passwd /etc/passwd
USER nobody

# Expose default port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/webhooked"]
CMD ["serve"]