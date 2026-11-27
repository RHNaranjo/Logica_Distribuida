# Proyecto Final Lógica Distribuida :)

Este proyecto incluye tablas de verdad, mundos posibles y deducción natural

## Arquitectura

```
Navegador ──:8000 (make abrir = port-forward)──▶ nginx-gateway
                                                  ├─ /      ──▶ frontend (:3000)
                                                  └─ /api/  ──▶ middleware (:8080)
                                                                  │ gRPC, round-robin por DNS headless
                      ┌───────────────────────────────────────────┼───────────────────────────┐
                      ▼                                           ▼                           ▼
             tablas ×3 pods (:9011)                     mundos ×3 pods (:9001)    deduccion ×3 pods (:9021)
                      └──────────────────────── publican el historial ────────────────────────┘
                                                                  ▼
                                          kafka :9092 (historial.<módulo>, 3 particiones)
                                                                  │ consumen (grupo historial-<módulo>)
                                                                  ▼
                                                           postgres (:5432)
```

Todo corre en minikube (una VM QEMU/KVM). Cada módulo es un Deployment con 3 réplicas.

| Componente                      | Qué hace                                                                                                  |
| ------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `nginx-gateway`                 | Única puerta: `/api/` al middleware y todo lo demás al frontend. Su `nginx.conf` vive en un ConfigMap.    |
| `frontend`                      | Sirve la página (React + Vite).                                                                           |
| `middleware`                    | Pone un id a cada petición (`X-Peticion`), registra el log y traduce HTTP a gRPC, balanceando entre pods. |
| `tablas`, `mundos`, `deduccion` | 3 pods c/u (servidores gRPC). Publican cada resolución en Kafka y la consumen para guardarla en Postgres. |
| `kafka`                         | Cola de mensajes (`apache/kafka:3.8.0`, KRaft). Un tópico por módulo con 3 particiones.                   |
| `postgres`                      | Base única `logica`, con PVC.                                                                             |

Service discovery: el DNS de Kubernetes. Cada módulo tiene un Service headless que resuelve a
las IPs de sus pods listos; el middleware balancea entre ellas del lado del cliente.

## Contrato gRPC

Los tres módulos cumplen el mismo servicio (`proto/modulo.proto`). La entrada y la salida son JSON libre (`google.protobuf.Struct`), así cada módulo conserva su formato:

```proto
service Modulo {
  rpc Resolver(google.protobuf.Struct) returns (google.protobuf.Struct);
  rpc Historial(PeticionHistorial) returns (google.protobuf.ListValue);
}
```

Además, cada instancia expone la salud estándar (`grpc.health.v1.Health`) y la reflexión de gRPC, para que `grpcurl` descubra los servicios sin el `.proto`.

El middleware abre una conexión `dns:///<módulo>:<puerto>` por módulo con la política
`round_robin`, y los módulos cierran cada conexión a los 30 s (`MaxConnectionAge`) para que
el cliente vuelva a resolver el DNS y vea los pods nuevos. Esto no viene del repositorio de
la clase: se basa en la documentación oficial de gRPC-Go.

Para regenerar `internal/rpc/` después de cambiar el `.proto`:

```bash
protoc --go_out=. --go_opt=module=logica --go-grpc_out=. --go-grpc_opt=module=logica proto/modulo.proto
```

## Cómo correrlo

Requisitos en Mac: `brew install qemu socket_vmnet minikube kubectl` y
`sudo brew services start socket_vmnet`. En Fedora: `libvirt` y `qemu-kvm`, y
`make cluster DRIVER=kvm2`.

```bash
make cluster      # la primera vez
make imagen       # construir la imagen dentro de minikube
make desplegar
make abrir        # se queda corriendo
```

Abrir `http://localhost:8000`. La pestaña **Inicio** muestra los pods de cada módulo, su
salud y cuántas peticiones ha atendido cada uno.

`instalar.sh` todavía usa docker compose; se actualiza al final de la etapa 6.

## Pruebas

Las etapas 1 a 4 se probaron con docker compose y Consul; esa versión está en el tag
`entrega-compose`.

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

### Etapa 6 — Kubernetes (minikube en una VM QEMU/KVM)

La versión con docker compose y Consul quedó guardada en el tag `entrega-compose`
(`git checkout entrega-compose` para volver a ella).

Requisitos en Mac (Apple Silicon): `brew install qemu socket_vmnet minikube kubectl`, y
`socket_vmnet` corriendo (`sudo brew services start socket_vmnet`). En Fedora se usa el
driver KVM: `make cluster DRIVER=kvm2` (necesita `libvirt` y `qemu-kvm`).

```bash
make cluster      # minikube en una VM con 6 GB y 4 CPUs
make imagen       # imagen logica:local dentro de la VM (desde el paso 2)
make desplegar    # Secret y ConfigMap desde .env y db/init.sql, y luego todo k8s/
```

Apagar y volver a prender el clúster conserva los datos de Postgres:

```bash
minikube stop
minikube start
```

#### Paso 1 — Postgres y Kafka en el clúster

Comprobar que el clúster es una VM:

```bash
minikube profile list
minikube ssh -- uname -m
```
Esperado: driver `qemu2` (o `kvm2`) y `aarch64` (o `x86_64` en Fedora).

Validar los manifiestos sin aplicarlos y esperar el despliegue:

```bash
kubectl apply -f k8s/ --dry-run=client
kubectl rollout status deployment/postgres
kubectl rollout status deployment/kafka
```
Esperado: `successfully rolled out` en ambos.

```bash
kubectl get deploy,rs,pods,svc,pvc
```
Esperado: 2 Deployments `1/1`, sus 2 ReplicaSets, 2 pods `Running` y el PVC `Bound`.

```bash
kubectl exec deploy/postgres -- sh -c 'psql -U $POSTGRES_USER -d logica -c "\dt"'
```
Esperado: la tabla `historial` (la creó `init.sql` desde el ConfigMap).

```bash
kubectl exec deploy/kafka -- /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka:9092 --create --topic prueba --partitions 3
kubectl exec deploy/kafka -- /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka:9092 --list
```
Esperado: `Created topic prueba.` y después `prueba`.

DNS entre pods (un pod temporal llega a Postgres por el nombre del Service):

```bash
kubectl run cliente --rm -i --restart=Never --image=postgres:16-alpine --env=PGPASSWORD=<DB_PASSWORD del .env> -- psql -h postgres -U <DB_USER del .env> -d logica -c 'select count(*) from historial'
```
Esperado: `0`.

Persistencia: insertar una fila, correr `kubectl delete pod -l app=postgres`, esperar el
pod nuevo (otro nombre; lo crea el ReplicaSet) y comprobar que la fila sigue ahí.

Prueba negativa: `kubectl delete secret postgres-secreto` y luego borrar el pod de
Postgres deja el pod nuevo en `CreateContainerConfigError`; `make config` lo arregla solo.

Limitaciones conocidas:
- Kafka no tiene disco: al reiniciar el clúster se pierden los tópicos de prueba y los
  mensajes pendientes. El historial importante vive en Postgres, que sí tiene PVC.
- Los Secrets de Kubernetes solo están codificados en base64, no cifrados.

#### Paso 2 — Los tres módulos, 3 réplicas cada uno

Cada módulo es un Deployment con `replicas: 3` (un ReplicaSet que mantiene 3 pods) y un
Service headless. Un pod = una instancia; su nombre es el del pod (Downward API).
Los módulos ya no se registran en Consul: el DNS del clúster los encuentra.

Construir la imagen dentro de minikube y desplegar:

```bash
make imagen
make desplegar
```

Después de cambiar código Go: `make imagen && make reiniciar`.

```bash
kubectl rollout status deployment/tablas
kubectl get deploy,rs
```
Esperado: `successfully rolled out`; `tablas`, `mundos` y `deduccion` en `3/3`, con un
ReplicaSet `3 3 3` cada uno. Repetir el `rollout status` con `mundos` y `deduccion`.

```bash
kubectl get endpoints tablas
```
Esperado: 3 direcciones `10.244.0.x:9011`. Si sale `<none>`, el `selector` del Service no
coincide con los `labels` del pod.

DNS headless (una IP por pod):

```bash
kubectl run dns --rm -i --restart=Never --image=busybox:1.36 -- nslookup tablas.default.svc.cluster.local
```
Esperado: tres `Address: 10.244.0.x`, las mismas del `get endpoints`.

gRPC directo a un pod (en otra terminal: `kubectl port-forward deploy/tablas 9011:9011`):

```bash
grpcurl -plaintext -d '{"formula": "p -> q"}' localhost:9011 logica.Modulo/Resolver
```
Esperado: la tabla con `"clasificacion": "contingencia"`. `port-forward` va a un solo pod.

Cada pod responde con su propio nombre (grpcurl dentro del clúster, contra cada IP):

```bash
for ip in $(kubectl get pods -l app=tablas -o jsonpath='{.items[*].status.podIP}'); do kubectl run grpc --rm -i --restart=Never --image=fullstorydev/grpcurl:latest -- -plaintext -v $ip:9011 grpc.health.v1.Health/Check | grep x-instancia; done
```
Esperado: tres `x-instancia: tablas-xxxxx-yyyyy`, cada uno con un pod distinto.

Historial y consumidores de Kafka (uno por pod, una partición cada uno):

```bash
kubectl exec deploy/postgres -- sh -c 'psql -U $POSTGRES_USER -d logica -c "select modulo, instancia, entrada from historial"'
kubectl exec deploy/kafka -- /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server kafka:9092 --describe --group historial-tablas --members
```
Esperado: la fila guardada con el nombre del pod en `instancia`, y 3 miembros con
`#PARTITIONS 1`.

Self-healing: `kubectl delete pod <pod-de-tablas>` y el ReplicaSet crea otro en segundos
(`kubectl get pods -l app=tablas -w`). Escalar: `kubectl scale deployment tablas
--replicas=5` (y de vuelta a 3; `make desplegar` también lo regresa, el YAML manda).

Prueba negativa: con Kafka apagado (`kubectl scale deployment kafka --replicas=0`), un pod
nuevo de tablas queda `0/1`, fuera de `kubectl get endpoints tablas`, y la liveness lo
reinicia a los ~50 s. Al restaurar Kafka (`--replicas=1`) pasa a `1/1` solo.

Limpiar las filas de prueba antes de la demo:

```bash
kubectl exec deploy/postgres -- sh -c 'psql -U $POSTGRES_USER -d logica -c "delete from historial"'
```

Notas:
- Desde este paso `docker compose` ya no funciona completo (los módulos no se registran en
  Consul). La versión con compose está en el tag `entrega-compose`.
- Tras `minikube start` los pods pueden verse `Unknown` o `0/1` hasta un minuto: es normal.
  Si `minikube start` falla con `Unable to query local socket address` (QEMU), correr
  `minikube stop` y luego `minikube start`.

#### Paso 3 — Middleware, frontend y gateway en el clúster

El sistema completo corre en minikube. Se borraron `docker-compose.yml`, `internal/consul/`
y la imagen propia de nginx (su `nginx.conf` pasó a un ConfigMap).

```bash
make imagen
make desplegar
make reiniciar
make abrir
```

Esperado: en `http://localhost:8000` el panel muestra `9/9 instancias sanas` y el recorrido
de cada respuesta es `nginx-gateway → middleware → <pod>`.

Revisar sin cambiar nada:

```bash
kubectl get deploy,pods,svc
kubectl logs deploy/middleware --tail=5
```
Esperado: `middleware`, `frontend` y `nginx-gateway` en `1/1`; los módulos en `3/3`;
`tablas`, `mundos` y `deduccion` con `CLUSTER-IP` en `None` (headless).

Rotación entre pods (balanceo del lado del cliente), en otra terminal:

```bash
for i in 1 2 3 4 5 6; do curl -si localhost:8000/api/tablas/salud | grep -i "^x-instancia"; done
```
Esperado: tres pods distintos que se repiten en orden.

Escalar:

```bash
kubectl scale deployment tablas --replicas=5
```
Esperado: el panel muestra los 5 pods y el tráfico llega a los nuevos en unos 30 s
(`MaxConnectionAge`). Regresar con `--replicas=3`.

Self-healing sin errores. Con este ciclo corriendo, borrar un pod desde otra terminal
(`kubectl delete pod <pod-de-tablas>`):

```bash
for i in $(seq 60); do curl -s -o /dev/null -w "%{http_code}\n" localhost:8000/api/tablas/salud; sleep 0.5; done | sort | uniq -c
```
Esperado: `60 200`.

Pruebas negativas:

```bash
curl -s -w " [%{http_code}]\n" localhost:8000/api/nada/salud
```
Esperado: `[404]` con `no existe el módulo nada`.

```bash
kubectl scale deployment tablas --replicas=0
curl -s -w " [%{http_code}]\n" -X POST localhost:8000/api/tablas/resolver -d '{"formula":"p"}'
kubectl scale deployment tablas --replicas=3
```
Esperado: `[503]` y la columna de `tablas` vacía en el panel; los otros módulos siguen. Al
volver a 3 tarda unos 15 s en responder `200`.

```bash
kubectl scale deployment middleware --replicas=0
curl -s -o /dev/null -w "%{http_code} en %{time_total}s\n" localhost:8000/api/estado
kubectl scale deployment middleware --replicas=1
```
Esperado: `504` en unos 3 s (`proxy_connect_timeout`) y la página sigue cargando.

Notas:
- Si cambias el `nginx.conf` del ConfigMap: `kubectl apply -f k8s/nginx-gateway.yaml` y
  `kubectl rollout restart deployment/nginx-gateway` (con `subPath`, nginx no ve el cambio solo).
- Si `make abrir` dice `address already in use`, algo más usa el puerto 8000 (por ejemplo,
  contenedores viejos de docker compose: `docker ps`).
- El middleware tiene una sola réplica: guarda en memoria cuántas peticiones atendió cada pod.

## Etapas

| Etapa | Cambio                                                           | Estado    |
| ----- | ---------------------------------------------------------------- | --------- |
| 1     | nginx como gateway HTTP (única puerta pública)                   | Listo     |
| 2     | gRPC entre middleware y módulos; nginx como balanceador gRPC     | Listo     |
| 3     | Consul para service discovery (registro, health check y borrado) | Listo     |
| 4     | Kafka para el historial (un tópico por módulo, consumo idempotente) | Listo     |
| 5     | Lógica de tablas de verdad, mundos posibles y deducción natural  | en curso  |
| 6     | Kubernetes (minikube en QEMU), mTLS y observabilidad             | en curso  |
