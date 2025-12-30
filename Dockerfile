# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o torrent-aio ./cmd/torrent-aio

# Runtime stage
FROM alpine:3.19

# Installer les dépendances runtime
RUN apk add --no-cache \
    mediainfo \
    ca-certificates \
    tzdata

# Créer un utilisateur non-root
RUN adduser -D -u 1000 aio
USER aio

WORKDIR /app

# Copier le binaire
COPY --from=builder /app/torrent-aio /usr/local/bin/torrent-aio

# Volume pour les fichiers à traiter
VOLUME ["/data"]

# Point d'entrée
ENTRYPOINT ["torrent-aio"]
CMD ["--help"]
