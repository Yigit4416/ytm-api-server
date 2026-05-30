# Build Stage
FROM golang:1.22-alpine AS builder

# We need gcc and musl-dev for SQLite CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum* ./
# Initialize go.sum if missing
RUN go mod tidy

COPY *.go ./
# Compile static binary with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o ytm-api-server .

# Runtime Stage
FROM alpine:latest

# Install Python, yt-dlp, and ffmpeg
RUN apk add --no-cache python3 py3-pip ffmpeg curl && \
    curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app
COPY --from=builder /app/ytm-api-server /app/ytm-api-server

# Expose API port
EXPOSE 8080

CMD ["/app/ytm-api-server"]
