#!/bin/bash

set -euo pipefail

if [[ "$#" -lt "2" ]]; then
	>&2 echo "Usage: openshift-4-destroyer.sh <prefix> <workflow>"
	exit 6
fi

PREFIX="${1}"
WORKFLOW_NAME="${2}"
CLUSTER_NAME="$(kubectl get workflow "${WORKFLOW_NAME}" -o yaml | yq '.metadata.labels["infra.stackrox.com/cluster-id"]')"

TIMESTAMP=$(date +%s)
RUNNER_NAME="${PREFIX}-${CLUSTER_NAME}-destroyer-${TIMESTAMP}"
AUTOMATION_FLAVORS_OS4_TAG="0.10.8"
OPENSHIFT_VERSION="ocp/stable"

PVC_NAME="${WORKFLOW_NAME}-data"
kubectl get pvc "${PVC_NAME}" >/dev/null || exit 1

manifest=$(cat <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: ${RUNNER_NAME}
spec:
  containers:
    - name: destroy
      image: quay.io/stackrox-io/ci:automation-flavors-openshift-4-${AUTOMATION_FLAVORS_OS4_TAG}
      imagePullPolicy: Always
      command:
        - entrypoint.sh
      args:
        - destroy
        - ${CLUSTER_NAME}
      env:
        - name: GOOGLE_CREDENTIALS
          valueFrom:
            secretKeyRef:
              name: openshift-4-gcp-service-account
              key: google-credentials.json
        - name: GCP_PROJECT
          value: "acs-team-temp-dev"
        - name: OPENSHIFT_VERSION
          value: "${OPENSHIFT_VERSION}"
      volumeMounts:
        - name: data
          mountPath: /data
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: ${PVC_NAME}
  restartPolicy: Never
EOF
)

echo "${manifest}" | kubectl apply -f -

sleep 20

kubectl logs -f "${RUNNER_NAME}"
