#!/bin/bash

echo "INFO: Download secrets"
ENVIRONMENT=development make secrets-download

echo "INFO: Place configuration with rendered values"
go run scripts/local-dev/main.go

echo "INFO: Replace /configuration/ -> ../../configuration/ in infra.yaml"
sed -i 's/\/configuration\//..\/..\/configuration\/tls\//g' configuration/infra.yaml
echo "INFO: Replace configuration/ -> ../../configuration/ in flavors.yaml"
sed -i 's/configuration\//..\/..\/configuration\//g' configuration/flavors.yaml

echo "INFO: Replace /etc/infra/static -> ../../ui/build in infra.yaml"
sed -i 's/\/etc\/infra\/static/..\/..\/ui\/build/g' configuration/infra.yaml

echo "INFO: Copy workflow templates in place"
cp chart/infra-server/static/{test,workflow}-*.yaml configuration/

echo "Prepare UI + CLI (for downloads)"
make -C ui
make cli
cp -R bin ui/build/downloads
