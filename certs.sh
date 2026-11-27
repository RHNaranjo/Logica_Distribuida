#!/usr/bin/env bash
# Generar la llave y un certificado por servicio

set -e

mkdir -p certs
cd certs

# La autoridad certificadora que firma todo, de 4096 bits
openssl genrsa -out ca.key 4096

# Certificación formato X509 por 10 años. el nombre común es LogicaCA
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=LogicaCA" \
  -addext "basicConstraints=critical,CA:TRUE"

# Un certificado por servicio. Su nombre está en el SAN (lista completa)
for s in tablas mundos deduccion middleware; do
  # Llave privada del servicio con 2048 bits
  openssl genrsa -out $s.key 2048

  # Petición para que el CA lo apruebe
  openssl req -new -key $s.key -out $s.csr -subj "/CN=$s"

  # Firmar la solicitud con vigencia de un año
  openssl x509 -req -days 365 -in $s.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out $s.crt \
    -extfile <(printf "subjectAltName=DNS:$s,DNS:$s.default.svc.cluster.local,DNS:localhost,IP:127.0.0.1")
done

# Borrar las solicitudes y el contador
rm *.csr *.srl
