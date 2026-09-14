# ── Etapa 1: compilar el frontend con Vite ──
FROM node:22-alpine AS web

WORKDIR /web

# Primero las dependencias: si package.json no cambió, Docker reutiliza
# esta capa y se salta el npm install completo
COPY frontend/package.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# ── Etapa 2: compilar los binarios de Go ──
FROM golang:1.26-alpine AS constructor

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/mundos ./cmd/mundos
RUN CGO_ENABLED=0 go build -o /bin/tablas ./cmd/tablas
RUN CGO_ENABLED=0 go build -o /bin/deduccion ./cmd/deduccion
RUN CGO_ENABLED=0 go build -o /bin/loadbalancer ./cmd/loadbalancer
RUN CGO_ENABLED=0 go build -o /bin/middleware ./cmd/middleware
RUN CGO_ENABLED=0 go build -o /bin/frontend ./cmd/frontend

# ── Etapa final: una sola imagen con todo ──
FROM alpine:3.21

WORKDIR /app

COPY --from=constructor /bin/mundos /app/mundos
COPY --from=constructor /bin/tablas /app/tablas
COPY --from=constructor /bin/deduccion /app/deduccion
COPY --from=constructor /bin/loadbalancer /app/loadbalancer
COPY --from=constructor /bin/middleware /app/middleware
COPY --from=constructor /bin/frontend /app/frontend

# El frontend ya compilado, listo para servirse
COPY --from=web /web/dist /app/web

# El docker-compose.yml elige cuál correr con `command:`
CMD ["/app/middleware"]
