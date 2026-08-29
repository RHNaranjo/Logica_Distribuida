FROM golang:1.26-alpine AS constructor

WORKDIR /src 

COPY go.mod go.sum ./ 

RUN go mod download 

COPY . .

# Multi-stage Dockerfile para que el contenedor no mate a la RAM
RUN CGO_ENABLED=0 go build -o /bin/coordinador ./cmd/coordinador/ 

RUN CGO_ENABLED=0 go build -o /bin/mundo ./cmd/mundo 

# Etapa final 
FROM alpine:3.21

WORKDIR /app 

COPY --from=constructor /bin/coordinador /app/coordinador

COPY --from=constructor /bin/mundo /app/mundo 

EXPOSE 8080

# El docker-compose.yml lo sobreescribirá
CMD ["/app/coordinador"]
