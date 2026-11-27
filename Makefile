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

# Generar los certificados en certs/ usando certs.sh
certificados:
	./certs.sh

# Crear o actualizar un Secret con el certificado de cada servicio
secretos:
	for s in tablas mundos deduccion middleware; do \
		kubectl create secret generic $$s-tls \
			--from-file=tls.crt=certs/$$s.crt \
			--from-file=tls.key=certs/$$s.key \
			--from-file=ca.crt=certs/ca.crt \
			--dry-run=client -o yaml | kubectl apply -f - ; \
	done

# Aplicar todos los archivos dentro de k8s/ para desplegar 
desplegar: config secretos
	kubectl apply -f k8s/

# Construir la imagen en minikube 
imagen:
	minikube image build -t logica:local .

# Recrear los pods que usan la imagen logica:local
reiniciar:
	kubectl rollout restart deployment/tablas deployment/mundos deployment/deduccion deployment/middleware deployment/frontend

# Abrir la página en http://localhost:8000 (se queda corriendo; Ctrl+C para cerrar)
abrir:
	kubectl port-forward svc/nginx-gateway 8000:8000
