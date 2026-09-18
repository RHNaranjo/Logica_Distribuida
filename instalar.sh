#!/usr/bin/env bash

# Para cuando falle el código
set -euo pipefail

# Trabajar siempre desde la carpeta del script, para que docker compose
# encuentre docker-compose.yml sin importar desde dónde se llame
cd "$(dirname "$0")"

# Colores para la terminal
VERDE='\033[0;32m'
AMARILLO='\033[1;33m'
SIN_COLOR='\033[0m'

# Log de algo correcto
log() {
  echo -e "${VERDE}==>${SIN_COLOR} $1"
}

# Log de error o advertencia
advertir() {
  echo -e "${AMARILLO}==> AVISO:${SIN_COLOR} $1"
}

# Verificar si estamos en fedora (Max dijo que tenía Fedora)
if [ ! -f /etc/os-release ] || ! grep -q '^ID=fedora' /etc/os-release; then
  echo "[ERROR] Esto sólo funciona para Fedora, que no se detectó en /etc/os-release"
  exit 1
fi

log "Fedora detectado. Comenzando la instalación!"

# Actualizar
sudo dnf upgrade --refresh -y

# Git, curl  y dnf-plugins-core (dnf config-manager)
sudo dnf install -y git curl dnf-plugins-core

# Docker Engine y COmpose
if command -v docker &>/dev/null; then
  log "Docker ya está instalado en esta computadora (versión: $(docker --version))."
else
  log "Instalando Docker Engine..."

  # Repositorio de Docker para Fedora
  sudo dnf config-manager addrepo --from-repofile=https://download.docker.com/linux/fedora/docker-ce.repo

  # Instalar el motor, la CLI, los contenedores, docker compose y Dockerfile multi-stage
  sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
fi

# Arrancar Docker ahora y en cada reinicio (fuera del "if": si Docker ya
# estaba instalado pero apagado, también hay que prenderlo)
sudo systemctl enable --now docker

# Instalar go
VERSION_GO="1.26.4"

if command -v go &>/dev/null && [[ "$(go version)" == *"go$VERSION_GO"* ]]; then
  log "Go ${VERSION_GO} ya está instalado."
else
  log "Instalando Go ${VERSION_GO}..."

  # Ver arquitectura de la computadora
  case "$(uname -m)" in
  x86_64) ARQUITECTURA_GO="amd64" ;;
  aarch64) ARQUITECTURA_GO="arm64" ;;
  *)
    echo "[ERROR] Arquitectura no soportada por este script: $(uname -m)"
    exit 1
    ;;
  esac

  TARBALL="go${VERSION_GO}.linux-${ARQUITECTURA_GO}.tar.gz"

  curl -fsSL "https://go.dev/dl/${TARBALL}" -o "/tmp/${TARBALL}"

  # Borrar instalaciones previas
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "/tmp/${TARBALL}"
  rm "/tmp/${TARBALL}"

  # Go para cualquier usuario en cualquier sesión de terminal
  echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null

  export PATH=$PATH:/usr/local/go/bin
fi

log "Versión de Go activa: $(go version)"

# Construir y levantar todo
log "Construyendo las imágenes (esto tarda varios minutos la primera vez)..."
sudo docker compose build

log "Levantando los 7 contenedores..."
sudo docker compose up -d

log "Listo. Verifica con: sudo docker compose ps"
log "Frontend en:    http://localhost:3000"
log "Middleware en:  http://localhost:8080/api/estado"
