# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o drift-guard ./cmd/server

# Final stage
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/drift-guard /drift-guard

EXPOSE 50051

ENTRYPOINT ["/drift-guard"]
