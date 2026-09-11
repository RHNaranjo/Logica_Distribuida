FROM golang:1.26-alpine AS constructor

WORKDIR /src 

COPY go.mod go.sum ./ 

RUN go mod download 

COPY . .

# Multi-stage Dockerfile para que el contenedor no mate a la RAM
RUN CGO_ENABLED=0 go build -o /bin/mundos ./cmd/mundos
RUN CGO_ENABLED=0 go build -o /bin/tablas ./cmd/tablas
RUN CGO_ENABLED=0 go build -o /bin/deduccion ./cmd/deduccion
RUN CGO_ENABLED=0 go build -o /bin/loadbalancer ./cmd/loadbalancer

# Etapa final
FROM alpine:3.21

WORKDIR /app

COPY --from=constructor /bin/mundos /app/mundos
COPY --from=constructor /bin/tablas /app/tablas
COPY --from=constructor /bin/deduccion /app/deduccion
COPY --from=constructor /bin/loadbalancer /app/loadbalancer

# El docker-compose.yml elige cuál correr después 
CMD ["/app/tablas"]
