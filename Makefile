cluster:
	minikube start --memory=6g --cpus=4

# Actualizar secret y configmap
# Crear o actualizar el Secret y el ConfigMap a partir de los archivos del repo
config:
	kubectl create secret generic postgres-secreto --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl create configmap postgres-init --from-file=db/init.sql --dry-run=client -o yaml | kubectl apply -f -
