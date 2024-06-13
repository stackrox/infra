#!/bin/bash

set -euo pipefail

if [[ "$#" -lt "2" ]]; then
	>&2 echo "Usage: gke-destroyer.sh <prefix> <workflow>"
	exit 6
fi

PREFIX="${1}"
WORKFLOW_NAME="${2}"
CLUSTER_NAME="$(kubectl get workflow "${WORKFLOW_NAME}" -o yaml | yq '.metadata.labels["infra.stackrox.com/cluster-id"]')"

TIMESTAMP=$(date +%s)
RUNNER_NAME="${PREFIX}-${CLUSTER_NAME}-destroyer-${TIMESTAMP}"
AUTOMATION_FLAVORS_TAG=$(yq '.annotations.automationFlavorsVersion' chart/infra-server/Chart.yaml)

manifest=$(cat <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: ${RUNNER_NAME}
spec:
  containers:
    - name: destroy
      image: quay.io/stackrox-io/ci:automation-flavors-gke-default-${AUTOMATION_FLAVORS_TAG}
      imagePullPolicy: Always
      command:
        - /usr/bin/entrypoint
      args:
        - destroy
        - --name=${CLUSTER_NAME}
        - --gcp-project=acs-team-temp-dev
      env:
        - name: GOOGLE_CREDENTIALS
          valueFrom:
            secretKeyRef:
              name: google-credentials
              key: google-credentials.json
      volumeMounts:
        - mountPath: /tmp
          name: credentials
  volumes:
    - name: credentials
      secret:
        defaultMode: 420
        secretName: google-credentials
  restartPolicy: Never
EOF
)

echo "${manifest}" | kubectl apply -f -

sleep 20

kubectl logs -f "${RUNNER_NAME}"
