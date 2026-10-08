# Proyecto Final Lógica Distribuida :) 


Este proyecto incluye tablas de verdad, mundos posibles y deducción natural 

## Arquitectura 
```
Navegador ──HTTP :8000──▶ nginx-gateway
                           ├─ /      ──▶ frontend (:3000)
                           └─ /api/  ──▶ middleware (:8080) ──▶ loadbalancer (:8081)
                                                                  ├─▶ tablas    ×3 (:9011-9013)
                                                                  ├─▶ mundos    ×3 (:9001-9003)
                                                                  └─▶ deduccion ×3 (:9021-9023)
                                                                          │
                                                                      postgres (:5432)
```

| Servicio | Qué hace |
|---|---|
| `nginx-gateway` | Manda `/api/` al middleware y todo lo demás al frontend. |
| `frontend` | Sirve la página (React + Vite). |
| `middleware` | Pone un id a cada petición (`X-Peticion`), registra el log y reenvía al balanceador. |
| `loadbalancer` | Registro de instancias, round-robin y health checks. |
| `tablas`, `mundos`, `deduccion` | Un contenedor por módulo, 3 instancias c/u. |
| `postgres` | Base única `logica`, compartida por las 9 instancias. |


## Cómo correrlo (Fedora para Max y Docker para todos)

Con Fedora nuevo:

```bash
bash instalar.sh
```

Con Docker ya instalado:

```bash
docker compose up --build -d
```

Abrir `http://localhost:8000`. La pestaña **Inicio** muestra el estado de las 9
instancias.

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
Esperado: `X-Nodo: nginx-gateway, middleware, loadbalancer`.

```bash
for i in 1 2 3 4 5 6; do curl -si localhost:8000/api/tablas/salud | grep -i x-instancia; done
```
Esperado: las instancias rotan (`tablas-0`, `tablas-1`, `tablas-2`, ...).

```bash
curl localhost:8080/salud
```
Esperado: `Connection refused`. El middleware ya no es público; solo se llega por nginx.

