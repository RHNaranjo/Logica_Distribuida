# Proyecto Final Lógica Distribuida :) 


Este proyecto incluye tablas de verdad, mundos posibles y deducción natural 

## Arquitectura 
```
Navegador ──HTTP :8000──▶ nginx-gateway
                           ├─ /      ──▶ frontend (:3000)
                           └─ /api/  ──▶ middleware (:8080)
                                              │ gRPC
                                              ▼
                          nginx-gateway :9000 (balanceador gRPC, round-robin)
                                  ├─▶ tablas    ×3 (:9011-9013)
                                  ├─▶ mundos    ×3 (:9001-9003)
                                  └─▶ deduccion ×3 (:9021-9023)
                                          │
                                      postgres (:5432)
```

| Servicio | Qué hace |
|---|---|
| `nginx-gateway` | En `:8000` manda `/api/` al middleware y todo lo demás al frontend. En `:9000` reparte las llamadas gRPC entre las instancias de cada módulo. |
| `frontend` | Sirve la página (React + Vite). |
| `middleware` | Pone un id a cada petición (`X-Peticion`), registra el log y traduce el HTTP del navegador a gRPC. |
| `tablas`, `mundos`, `deduccion` | Un contenedor por módulo, 3 instancias c/u (servidores gRPC). |
| `postgres` | Base única `logica`, compartida por las 9 instancias. |

## Contrato gRPC

Los tres módulos cumplen el mismo servicio (`proto/modulo.proto`). La entrada y la salida son JSON libre (`google.protobuf.Struct`), así cada módulo conserva su formato:

```proto
service Modulo {
  rpc Resolver(google.protobuf.Struct) returns (google.protobuf.Struct);
  rpc Historial(PeticionHistorial) returns (google.protobuf.ListValue);
}
```

Además, cada instancia expone la salud estándar (`grpc.health.v1.Health`) y la reflexión de gRPC, para que `grpcurl` descubra los servicios sin el `.proto`.

nginx elige el módulo con el header `modulo` de cada llamada. La configuración de gRPC no viene del repositorio de la clase: se basa en la documentación oficial de nginx (`ngx_http_grpc_module`) y de gRPC-Go.

Para regenerar `internal/rpc/` después de cambiar el `.proto`:

```bash
protoc --go_out=. --go_opt=module=logica --go-grpc_out=. --go-grpc_opt=module=logica proto/modulo.proto
```

## Cómo correrlo (Fedora para Max y Docker para todos)

Con Fedora nuevo:

```bash
bash instalar.sh
```

Con Docker ya instalado:

```bash
docker compose up --build -d
```

Abrir `http://localhost:8000`. La pestaña **Inicio** muestra cuántas peticiones ha atendido cada instancia (aparecen conforme responden).

## Pruebas

### Etapa 1 — nginx como gateway

```bash
docker compose exec nginx-gateway nginx -t
```
Esperado: `syntax is ok` y `test is successful`.

```bash
curl localhost:8000/salud
```
Esperado: `nginx-gateway vivo`.

```bash
curl -si localhost:8000/api/estado | grep -i x-nodo
```
Esperado (desde la etapa 2): `X-Nodo: nginx-gateway, middleware`.

```bash
curl localhost:8080/salud
```
Esperado: `Connection refused`. El middleware ya no es público; solo se llega por nginx.

### Etapa 2 — gRPC y nginx como balanceador

Requiere `grpcurl`: `brew install grpcurl` en Mac, o
`go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest` en Fedora.

```bash
grpcurl -plaintext -H 'modulo: tablas' localhost:9000 list
```
Esperado: `grpc.health.v1.Health`, `grpc.reflection.v1.ServerReflection`,
`grpc.reflection.v1alpha.ServerReflection` y `logica.Modulo`.

```bash
for i in 1 2 3 4 5 6; do grpcurl -plaintext -v -H 'modulo: tablas' localhost:9000 grpc.health.v1.Health/Check | grep x-instancia; done
```
Esperado: `x-instancia: tablas-0`, `tablas-1`, `tablas-2`, y se repite (round-robin de nginx).

```bash
grpcurl -plaintext -H 'modulo: mundos' -d '{"formula": "[]p"}' localhost:9000 logica.Modulo/Resolver
```
Esperado: `Code: Unimplemented` con el mensaje `mundos posibles: ... no ha sido implementado`
(la lógica llega en la etapa 5).

```bash
grpcurl -plaintext -H 'modulo: tablas' -d '{"limite": 3}' localhost:9000 logica.Modulo/Historial
```
Esperado: una lista con hasta 3 registros del historial de `tablas`.

```bash
grpcurl -plaintext -H 'modulo: nada' localhost:9000 list
```
Esperado: un error `Unavailable` con `502 (Bad Gateway)`. nginx no conoce ese módulo.

```bash
curl -si -X POST localhost:8000/api/tablas/resolver -d '{"formula":"p"}' | grep -iE "^HTTP|x-instancia|x-nodo"
```
Esperado: `HTTP/1.1 501`, `X-Instancia: tablas-N` y
`X-Nodo: nginx-gateway, middleware, nginx-lb, tablas-N`. El navegador sigue en HTTP;
el middleware hizo la llamada por gRPC.

## Etapas

| Etapa | Cambio | Estado |
|---|---|---|
| 1 | nginx como gateway HTTP (única puerta pública) | ✅ |
| 2 | gRPC entre middleware y módulos; nginx como balanceador gRPC | ✅ |
| 3 | Consul para service discovery | pendiente |
| 4 | Kafka para el historial | pendiente |
| 5 | Lógica de tablas de verdad y mundos posibles | pendiente |
