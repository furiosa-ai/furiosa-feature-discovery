FROM golang:1.27.2-bookworm AS builder

WORKDIR /workspace
COPY . .

# Build feature-discovery binary
RUN CGO_ENABLED=0 go build -o /furiosa-feature-discovery cmd/main.go

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /opt/bin

COPY --from=builder /furiosa-feature-discovery /opt/bin/furiosa-feature-discovery

CMD ["./furiosa-feature-discovery"]
