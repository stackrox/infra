#!/bin/bash

# This script finds infra cluster names for obscure VM prefixes from the Janitor control output
# gcloud compute instances list --project acs-team-temp-dev --format json | jq -r '.[].name' | sed 's/gke-//; s/-default-pool.*//; s/-master.*//; s/-worker.*//' | sort | uniq

set -euo pipefail

if [[ "$#" -lt "1" ]]; then
	>&2 echo "Usage: find-demo-clusters-for-vms.sh <VM prefix>"
	exit 6
fi

CLUSTER="$1"
INSTANCE=$(gcloud compute instances list --project acs-team-temp-dev --format json \
  | jq -r '.[].name' \
  | grep "^gke-${CLUSTER}.*" \
  | head -n 1)

gcloud compute instances describe "${INSTANCE}" --project acs-team-temp-dev --format json | jq -r '.labels.name'
