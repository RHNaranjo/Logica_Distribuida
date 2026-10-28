# Proyecto Final Lógica Distribuida :)

Este proyecto incluye tablas de verdad, mundos posibles y deducción natural

## Arquitectura

```
Navegador ──HTTP :8000──▶ nginx-gateway
                           ├─ /      ──▶ frontend (:3000)
                           └─ /api/  ──▶ middleware (:8080) ──▶ Consul (:8500)
                                              │ gRPC               ▲  registro y salud
                                              ▼                    │
                          nginx-gateway :9000 (balanceador gRPC)   │
                                  ├─▶ tablas    ×3 (:9011-9013) ───┤
                                  ├─▶ mundos    ×3 (:9001-9003) ───┤
                                  └─▶ deduccion ×3 (:9021-9023) ───┘
                                       │ publican el historial
                                       ▼
                                  kafka :9092 (historial.<módulo>, 3 particiones)
                                       │ consumen (grupo historial-<módulo>)
                                       ▼
                                  postgres (:5432)
```

| Servicio                        | Qué hace                                                                                                                                                                                                           |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `nginx-gateway`                 | En `:8000` manda `/api/` al middleware y todo lo demás al frontend. En `:9000` reparte las llamadas gRPC entre las instancias sanas que Consul le indica (consul-template regenera su configuración y lo recarga). |
| `consul`                        | Service discovery: las instancias se registran y se borran solas, y Consul les hace health check gRPC cada 5 s. Página en `:8500/ui`.                                                                              |
| `frontend`                      | Sirve la página (React + Vite).                                                                                                                                                                                    |
| `middleware`                    | Pone un id a cada petición (`X-Peticion`), registra el log y traduce el HTTP del navegador a gRPC.                                                                                                                 |
| `tablas`, `mundos`, `deduccion` | Un contenedor por módulo, 3 instancias c/u (servidores gRPC). Publican cada resolución en su tópico de Kafka y la consumen para guardarla en Postgres.                                                                                                                                                      |
| `kafka`                         | Cola de mensajes (`apache/kafka:3.8.0`, KRaft). Un tópico por módulo con 3 particiones. El historial se escribe de forma asíncrona. |
| `postgres`                      | Base única `logica`, compartida por las 9 instancias.                                                                                                                                                              |

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

Abrir `http://localhost:8000`. La pestaña **Inicio** muestra las instancias que Consul
tiene registradas, su salud y cuántas peticiones ha atendido cada una. La página de
Consul está en `http://localhost:8500/ui`.

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

### Etapa 3 — Consul como service discovery

```bash
curl -s localhost:8500/v1/catalog/services
```

Esperado: `{"consul": [], "deduccion": [], "mundos": [], "tablas": []}` (impreso con saltos de línea).

```bash
curl -s localhost:8500/v1/health/state/passing | grep -o '"ServiceID": *"[a-z]*-[0-9]"'
```

Esperado: 9 líneas, de `deduccion-0` a `tablas-2`.

```bash
docker compose exec nginx-gateway sed -n '/upstream tablas/,/}/p' /etc/nginx/nginx.conf
```

Esperado: tres líneas `server tablas:9011`, `:9012` y `:9013`, generadas por consul-template.

```bash
curl -s localhost:8000/api/estado
```

Esperado: las 9 instancias con `"sana":true`.

```bash
docker compose stop tablas && sleep 3 && curl -s localhost:8500/v1/catalog/services
```

Esperado: `tablas` ya no aparece (la instancia se borró al apagarse). En `/api/estado`
tampoco, y `nginx.conf` muestra `server 127.0.0.1:65535 down;` en `upstream tablas`.

```bash
grpcurl -plaintext -H 'modulo: tablas' localhost:9000 list
```

Esperado (con `tablas` apagado): `Unavailable` con `502 (Bad Gateway)`. Los demás módulos
siguen respondiendo.

```bash
docker compose start tablas && sleep 12 && curl -s localhost:8000/api/estado
```

Esperado: `tablas-0`, `tablas-1` y `tablas-2` regresan solas con `"sana":true`.

```bash
docker compose stop tablas && docker compose restart nginx-gateway && docker compose logs nginx-gateway | tail -3
```

Esperado: nginx arranca sin error con `tablas` apagado (en la etapa 2 no lo hacía).

```bash
docker compose kill tablas && sleep 7 && curl -s localhost:8000/api/estado
```

Esperado: las instancias de `tablas` salen con `"sana":false`. Consul las borra solo
después de 1 minuto (`DeregisterCriticalServiceAfter`).

### Etapa 4 — Kafka para el historial

`db/init.sql` cambió (columna `evento_id`), así que la base se crea de nuevo. **Esto
borra el historial anterior:**

```bash
docker compose down -v && docker compose up --build -d --remove-orphans
```
Esperado: 9 servicios en `Up`, incluido `kafka`.

```bash
docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
```
Esperado: `historial.deduccion`, `historial.mundos`, `historial.tablas` (y `__consumer_offsets`).

```bash
docker compose exec kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic historial.tablas
```
Esperado: `PartitionCount: 3`.

```bash
for i in 1 2 3; do curl -s -o /dev/null -w "%{http_code}\n" -X POST localhost:8000/api/tablas/resolver -d "{\"formula\":\"p$i\"}"; done
```
Esperado: `501` tres veces (la lógica llega en la etapa 5, pero el intento sí se anota).

```bash
docker compose exec kafka /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic historial.tablas --from-beginning --max-messages 3
```
Esperado: 3 eventos en JSON, cada uno con su `evento_id` y su `instancia`.

```bash
docker compose exec kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group historial-tablas
```
Esperado: 3 particiones y `LAG` en `0`: el consumidor ya guardó todo.

```bash
curl -s "localhost:8000/api/tablas/historial?limite=3"
```
Esperado: los 3 registros, leídos de Postgres.

**Resiliencia:** con Postgres apagado, el usuario no se entera y Kafka guarda los eventos.

```bash
docker compose stop postgres
for i in 4 5 6; do curl -s -o /dev/null -w "%{http_code}\n" -X POST localhost:8000/api/tablas/resolver -d "{\"formula\":\"q$i\"}"; done
docker compose exec kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group historial-tablas
```
Esperado: `501` al instante, y el `LAG` sumando `3` entre las particiones.

```bash
docker compose start postgres && sleep 15 && curl -s "localhost:8000/api/tablas/historial?limite=10"
```
Esperado: `LAG` de nuevo en `0` y los registros `q4`, `q5` y `q6` en el historial.

**Idempotencia:** reenviar un evento ya guardado no lo duplica.

```bash
curl -s "localhost:8000/api/tablas/historial?limite=200" | grep -o '"modulo"' | wc -l
docker compose exec kafka sh -c "/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic historial.tablas --from-beginning --max-messages 1 | /opt/kafka/bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic historial.tablas"
sleep 3 && curl -s "localhost:8000/api/tablas/historial?limite=200" | grep -o '"modulo"' | wc -l
```
Esperado: el mismo número antes y después (`ON CONFLICT (evento_id) DO NOTHING`).

**Límite conocido:** si Kafka está apagado, el usuario recibe su respuesta en ≤3 s, pero
ese evento no se guarda (`docker compose logs tablas` muestra
`No se pudo publicar el historial`). Al volver Kafka, el consumidor tarda unos segundos
en reunirse al grupo y sigue guardando.

### Etapa 5 — Lógica: tablas de verdad

```bash
go test ./internal/modulos/tablas/ -v
```
Esperado: `PASS` en `TestResolver` y `TestResolverErrores`.

```bash
curl -s -X POST localhost:8000/api/tablas/resolver -d '{"formula":"p -> q"}'
```
Esperado: `"clasificacion":"contingencia"`, `"variables":["p","q"]` y 4 renglones; el
único falso es `[true,false]`.

```bash
curl -s -X POST localhost:8000/api/tablas/resolver -d '{"formula":"p | ~p"}'
curl -s -X POST localhost:8000/api/tablas/resolver -d '{"formula":"p & ~p"}'
```
Esperado: `tautologia` y `contradiccion`.

```bash
curl -s -w " [%{http_code}]\n" -X POST localhost:8000/api/tablas/resolver -d '{"formula":"[]p"}'
curl -s -w " [%{http_code}]\n" -X POST localhost:8000/api/tablas/resolver -d '{"formula":"a & b & c & d & e & f & g & h & i & j & k"}'
```
Esperado: `[400]` con el motivo (fórmula modal; más de 10 variables).

```bash
grpcurl -plaintext -H 'modulo: tablas' -d '{"formula": "p | ~p"}' localhost:9000 logica.Modulo/Resolver
```
Esperado: la misma tabla por gRPC, con `"clasificacion": "tautologia"`.

## Etapas

| Etapa | Cambio                                                           | Estado    |
| ----- | ---------------------------------------------------------------- | --------- |
| 1     | nginx como gateway HTTP (única puerta pública)                   | Listo     |
| 2     | gRPC entre middleware y módulos; nginx como balanceador gRPC     | Listo     |
| 3     | Consul para service discovery (registro, health check y borrado) | Listo     |
| 4     | Kafka para el historial (un tópico por módulo, consumo idempotente) | Listo     |
| 5     | Lógica de tablas de verdad y mundos posibles                     | en curso |
