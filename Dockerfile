# --- Stage 1: The Builder ---
FROM golang:1.25.6-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o portfolio main.go

# --- Stage 2: The Secure Runner ---
FROM alpine:latest

# Install SSH tools
RUN apk add --no-cache openssh-client ca-certificates

# 1. Create User (UID 1000)
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -D appuser

# 2. Set the Environment Variable (THIS IS THE FIX)
ENV HOME=/home/appuser

WORKDIR /home/appuser

# 3. Copy binary
COPY --from=builder /app/portfolio .

# 4. Create SSH folder & permissions
RUN mkdir .ssh && chown -R appuser:appgroup /home/appuser

# 5. Env Vars
ENV TERM=xterm-256color
ENV COLORTERM=truecolor

# 6. Switch User
USER appuser

EXPOSE 23234
CMD ["./portfolio"]