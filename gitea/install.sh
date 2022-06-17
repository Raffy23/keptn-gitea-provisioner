#!/bin/bash

NAMESPACE=default

kubectl apply -n ${NAMESPACE} -f admin-credentials.yaml

helm install gitea gitea-charts/gitea \
	--set memcached.enabled=false \
	--set postgresql.enabled=false \
	--set gitea.config.database.DB_TYPE=sqlite3 \
	--set gitea.admin.existingSecret=gitea-admin-secret \
	--set gitea.config.server.OFFLINE_MODE=true \
	--set gitea.config.server.ROOT_URL=http://gitea-http.${NAMESPACE}:3000/

