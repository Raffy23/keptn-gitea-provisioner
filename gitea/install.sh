#!/bin/bash
NAMESPACE=default

# Apply the admin credentials for the gitea and the provisioner service
kubectl apply -n ${NAMESPACE} -f admin-credentials.yaml

# Add the gitea helm charts and install gitea to the cluster
helm repo add gitea-charts https://dl.gitea.io/charts/
helm repo update
helm install gitea gitea-charts/gitea \
	--set memcached.enabled=false \
	--set postgresql.enabled=false \
	--set gitea.config.database.DB_TYPE=sqlite3 \
	--set gitea.admin.existingSecret=gitea-admin-secret \
	--set gitea.config.server.OFFLINE_MODE=true \
	--set gitea.config.server.ROOT_URL=http://gitea-http.${NAMESPACE}:3000/
