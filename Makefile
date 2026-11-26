# qemu para Mac y kvm2 para Fedora 
DRIVER ?= qemu

# Arrancar en VM
cluster:
ifeq ($(DRIVER),qemu)
	minikube start --driver=qemu --network=socket_vmnet --memory=6g --cpus=4
else 
	minikube start --driver=$(DRIVER) --memory=6g --cpus=4
endif

# Actualizar secret y configmap
# Crear o actualizar el Secret y el ConfigMap a partir de los archivos del repo
config:
	kubectl create secret generic postgres-secreto --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl create configmap postgres-init --from-file=db/init.sql --dry-run=client -o yaml | kubectl apply -f -

# Aplicar todos los archivos dentro de k8s/ y desplegar 
desplegar: config 
	kubectl apply -f k8s/

# Construir la imagen en minikube 
imagen:
	minikube image build -t logica:local .

# Recrear los pods de los módulos con una nueva imagen 
reiniciar:
	kubectl rollout restart deployment/tablas deployment/mundos deployment/deduccion 

