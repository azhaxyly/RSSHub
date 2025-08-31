# ---- builder ----
FROM golang:1.24.3 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rsshub ./cmd

# ---- runner ----
FROM debian:stable-slim AS runner
RUN apt-get update \
&& apt-get install -y --no-install-recommends ca-certificates tzdata \
&& rm -rf /var/lib/apt/lists/* \
&& update-ca-certificates
WORKDIR /app
COPY --from=builder /app/rsshub .
CMD ["./rsshub", "fetch"]