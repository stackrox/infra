# Migration Plan: GKE to OpenShift Container Platform 4

## Context

The infra-server application is currently deployed to GKE (Google Kubernetes Engine) clusters in GCP us-west2 region using Helm charts. This migration plan addresses the need to move the deployment to a standard OpenShift Container Platform 4 cluster while maintaining the existing Helm-based deployment approach.

**Current State:**
- Deployed via Helm chart located at `chart/infra-server/`
- Uses GKE-specific resources: ManagedCertificate, Ingress with global static IPs
- Integrates with GCP services: Secret Manager, GCS for Argo artifacts
- Two environments: development (dev.infra.rox.systems) and production (infra.rox.systems)
- Dependencies: argo-workflows (v1.0.4), kube-prometheus-stack (v83.5.1)

**Migration Goal:**
Direct migration from GKE to OCP 4 (cluster already provisioned) using:
- External Secrets Operator for secret management (continue using GCP Secret Manager)
- GCS for Argo Workflows artifact storage (no migration needed)
- OpenShift Routes for ingress with built-in TLS

## Implementation Approach

The migration follows a **conditional templating pattern** where GKE-specific resources are disabled and OCP 4 equivalents are rendered based on a `platform` value. This leverages the existing multi-cloud support structure (aws/, azure/, aro/, etc. secret templates) and extends it to the deployment layer.

### Key Changes Required

1. **Ingress & TLS**: Replace GKE ManagedCertificate + Ingress with OpenShift Routes
2. **Service**: Change from NodePort to ClusterIP (Routes handle external access)
3. **Security**: Add SecurityContext compatible with OCP restricted-v2 SCC
4. **Secrets**: Deploy External Secrets Operator to sync from GCP Secret Manager
5. **Deployment Scripts**: Create OCP 4 variant of helm.sh using `oc` instead of `kubectl`
6. **CI/CD**: New GitHub Actions workflow for OCP 4 authentication and deployment

## Critical Files to Modify

### 1. Helm Chart Templates

**Create new OCP 4 route template:**
- `chart/infra-server/templates/ocp4/route.yaml` - OpenShift Route resource

**Modify existing templates to be platform-conditional:**
- `chart/infra-server/templates/ingress.yaml` - Wrap in `{{- if ne .Values.platform "ocp4" -}}`
- `chart/infra-server/templates/certificate.yaml` - Wrap in `{{- if ne .Values.platform "ocp4" -}}`
- `chart/infra-server/templates/service.yaml` - Conditional type and annotations
- `chart/infra-server/templates/deployment.yaml` - Add OCP 4 securityContext

**Create External Secrets resources:**
- `chart/infra-server/templates/ocp4/secret-store.yaml` - ClusterSecretStore for GCP Secret Manager
- `chart/infra-server/templates/ocp4/external-secrets.yaml` - ExternalSecret resources

### 2. Configuration Files

**Create:**
- `chart/infra-server/ocp4-values.yaml` - OCP 4-specific values (platform, route config, security)
- `chart/infra-server/configuration/ocp4-development-values.yaml` - Dev environment config
- `chart/infra-server/configuration/ocp4-production-values.yaml` - Prod environment config

### 3. Deployment Automation

**Create:**
- `scripts/deploy/ocp4-helm.sh` - OCP 4 deployment script using `oc` commands

**Update:**
- `.github/workflows/deploy-ocp4.yaml` - GitHub Actions workflow for OCP 4
- `Makefile` - Add ocp4-helm-template, ocp4-helm-deploy, ocp4-helm-diff targets

## Detailed Implementation Steps

### Step 1: Create OpenShift Route Template

Create `chart/infra-server/templates/ocp4/route.yaml`:

```yaml
{{- if and (eq .Values.testMode false) (eq .Values.platform "ocp4") -}}
---
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: infra-server-route-primary
  namespace: infra
  annotations:
    haproxy.router.openshift.io/timeout: 5m
spec:
  host: {{ .Values.hosts.primary }}
  tls:
    termination: edge
    insecureEdgeTerminationPolicy: Redirect
  to:
    kind: Service
    name: infra-server-service
    weight: 100
  port:
    targetPort: https
---
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: infra-server-route-secondary
  namespace: infra
  annotations:
    haproxy.router.openshift.io/timeout: 5m
spec:
  host: {{ .Values.hosts.secondary }}
  tls:
    termination: edge
    insecureEdgeTerminationPolicy: Redirect
  to:
    kind: Service
    name: infra-server-service
    weight: 100
  port:
    targetPort: https
{{- end }}
```

### Step 2: Make Existing Templates Platform-Conditional

**Update `chart/infra-server/templates/ingress.yaml`:**
```yaml
{{- if and (eq .Values.testMode false) (ne .Values.platform "ocp4") -}}
# ... existing ingress content
{{ end }}
```

**Update `chart/infra-server/templates/certificate.yaml`:**
```yaml
{{- if and (eq .Values.testMode false) (ne .Values.platform "ocp4") -}}
# ... existing ManagedCertificate content
{{ end }}
```

**Update `chart/infra-server/templates/service.yaml`:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: infra-server-service
  namespace: infra
  labels:
    app.kubernetes.io/name: infra-server
  annotations:
{{- if eq .Values.platform "ocp4" }}
    service.beta.openshift.io/serving-cert-secret-name: infra-server-serving-cert
{{- else }}
    cloud.google.com/app-protocols: '{"https":"HTTP2"}'
{{- end }}
spec:
{{- if eq .Values.platform "ocp4" }}
  type: ClusterIP
{{- else }}
  type: NodePort
{{- end }}
  selector:
    app.kubernetes.io/name: infra-server
  ports:
    - protocol: TCP
      port: 8443
      targetPort: 8443
      name: https
```

**Update `chart/infra-server/templates/deployment.yaml`:**

Add securityContext for OCP 4:
```yaml
spec:
  template:
    spec:
{{- if eq .Values.platform "ocp4" }}
      securityContext:
        runAsNonRoot: true
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault
{{- end }}
      containers:
        - name: infra-server
          # ... existing container config
{{- if eq .Values.platform "ocp4" }}
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop:
              - ALL
            seccompProfile:
              type: RuntimeDefault
{{- end }}
```

### Step 3: Create External Secrets Configuration

**Create `chart/infra-server/templates/ocp4/secret-store.yaml`:**
```yaml
{{- if eq .Values.platform "ocp4" -}}
---
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: gcpsm-secret-store
spec:
  provider:
    gcpsm:
      projectID: acs-team-automation
      auth:
        secretRef:
          secretAccessKeySecretRef:
            name: gcpsm-secret
            key: secret-access-credentials
            namespace: external-secrets-system
{{- end }}
```

**Create `chart/infra-server/templates/ocp4/external-secrets.yaml`:**
```yaml
{{- if eq .Values.platform "ocp4" -}}
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: infra-server-external-secret
  namespace: infra
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: gcpsm-secret-store
    kind: ClusterSecretStore
  target:
    name: infra-server-secrets
    creationPolicy: Owner
    template:
      type: Opaque
      data:
        google-credentials.json: "{{ `{{ .googleCredentials }}` }}"
        bigquery-sa.json: "{{ `{{ .bigquerySa }}` }}"
        oidc.yaml: "{{ `{{ .oidc }}` }}"
        cert.pem: "{{ `{{ .cert }}` }}"
        key.pem: "{{ `{{ .key }}` }}"
        infra.yaml: "{{ `{{ .infra }}` }}"
  dataFrom:
  - extract:
      key: infra-values-{{ .Values.environment }}
  - extract:
      key: infra-values-from-files-{{ .Values.environment }}
{{- end }}
```

### Step 4: Create OCP 4 Values Files

**Create `chart/infra-server/ocp4-values.yaml`:**
```yaml
# Platform identifier
platform: ocp4

# OpenShift-specific configuration
ocp4:
  enabled: true
  route:
    enabled: true
    tls:
      termination: edge
      insecureEdgeTerminationPolicy: Redirect
  securityContext:
    enabled: true
    runAsNonRoot: true
    fsGroup: 1000

# Service configuration for OCP
service:
  type: ClusterIP

# Storage class (adjust based on OCP cluster configuration)
storageClassName: gp3-csi
```

### Step 5: Create OCP 4 Deployment Script

**Create `scripts/deploy/ocp4-helm.sh`:**

```bash
#!/usr/bin/env bash
set -euo pipefail

TASK="$1"
TAG="$2"
ENVIRONMENT="$3"
SECRET_VERSION="${4:-latest}"

TEST_MODE="${TEST_MODE:-false}"
HELM_MONITORING_FINAL_SET=()
if [[ "${NO_MONITORING:-false}" == "true" ]]; then
    HELM_MONITORING_FINAL_SET=(--set monitoring.enabled=false)
fi

SECRETS_PROJECT="acs-team-automation"
RELEASE_NAMESPACE="infra"
RELEASE_NAME="infra-server"

check_not_empty() {
    for V in "$@"; do
        typeset -n VAR="$V"
        if [ -z "${VAR:-}" ]; then
            echo "ERROR: Variable $V is not set or empty"
            exit 1
        fi
    done
}

install_crds() {
    argo_chart_file=$(find "chart/infra-server/charts" -name "argo-workflows-*.tgz" 2>/dev/null | head -1)
    ARGO_WORKFLOWS_APP_VERSION="$(tar -xzOf "${argo_chart_file}" argo-workflows/Chart.yaml | yq eval '.appVersion' -)"
    echo "Using argo-workflows app version: ${ARGO_WORKFLOWS_APP_VERSION}" >&2
    oc apply --kustomize \
        "https://github.com/argoproj/argo-workflows/manifests/base/crds/minimal?ref=${ARGO_WORKFLOWS_APP_VERSION}" >&2
}

template() {
    helm upgrade \
        "${RELEASE_NAME}" \
        chart/infra-server \
        --install \
        --create-namespace \
        --dry-run \
        --namespace "${RELEASE_NAMESPACE}" \
        --values chart/infra-server/argo-values.yaml \
        --values chart/infra-server/monitoring-values.yaml \
        --values chart/infra-server/ocp4-values.yaml \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set testMode="${TEST_MODE}" \
        --values - \
        "${HELM_MONITORING_FINAL_SET[@]}" \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    )
}

deploy() {
    helm upgrade \
        "${RELEASE_NAME}" \
        chart/infra-server \
        --install \
        --create-namespace \
        --timeout 5m \
        --wait \
        --namespace "${RELEASE_NAMESPACE}" \
        --values chart/infra-server/argo-values.yaml \
        --values chart/infra-server/monitoring-values.yaml \
        --values chart/infra-server/ocp4-values.yaml \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set testMode="${TEST_MODE}" \
        --values - \
        "${HELM_MONITORING_FINAL_SET[@]}" \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    )
}

diff() {
    helm upgrade \
        "${RELEASE_NAME}" \
        chart/infra-server \
        --install \
        --create-namespace \
        --dry-run \
        --namespace "${RELEASE_NAMESPACE}" \
        --values chart/infra-server/argo-values.yaml \
        --values chart/infra-server/monitoring-values.yaml \
        --values chart/infra-server/ocp4-values.yaml \
        --set tag="${TAG}" \
        --set environment="${ENVIRONMENT}" \
        --set testMode="${TEST_MODE}" \
        --values - \
        "${HELM_MONITORING_FINAL_SET[@]}" \
    < <(gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    && gcloud secrets versions access "${SECRET_VERSION}" \
        --secret "infra-values-from-files-${ENVIRONMENT}" \
        --project "${SECRETS_PROJECT}" \
    ) \
    | sed -n '/---/,$p' \
    | oc diff -R -f -
}

check_not_empty TASK TAG ENVIRONMENT
install_crds
eval "$TASK"
```

Make executable:
```bash
chmod +x scripts/deploy/ocp4-helm.sh
```

### Step 6: Update Makefile

Add to `Makefile`:

```makefile
# OCP 4 deployment targets
.PHONY: ocp4-helm-template
ocp4-helm-template: helm-dependency-update
	@./scripts/deploy/ocp4-helm.sh template $(VERSION) $(ENVIRONMENT) $(SECRET_VERSION)

.PHONY: ocp4-helm-deploy
ocp4-helm-deploy: helm-dependency-update
	@./scripts/deploy/ocp4-helm.sh deploy $(VERSION) $(ENVIRONMENT) $(SECRET_VERSION)

.PHONY: ocp4-helm-diff
ocp4-helm-diff: helm-dependency-update
	@./scripts/deploy/ocp4-helm.sh diff $(VERSION) $(ENVIRONMENT) $(SECRET_VERSION)
```

### Step 7: Create GitHub Actions Workflow

**Create `.github/workflows/deploy-ocp4.yaml`:**

```yaml
name: Deploy infra to OCP 4
run-name: >-
  ${{
    format('Deploy infra version {0} to OCP 4 {1}',
      inputs.version,
      inputs.environment
    )
  }}

on:
  workflow_dispatch:
    inputs:
      environment:
        description: Dev or Prod?
        required: true
        default: development
        type: choice
        options:
        - development
        - production
      version:
        description: Version, expanded to Github + Docker image tag
        required: true

jobs:
  wait-for-images:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        image: [infra-server, infra-certifier]
    steps:
      - name: Wait for image
        uses: stackrox/actions/release/wait-for-image@v1
        with:
          token: ${{ secrets.QUAY_RHACS_ENG_BEARER_TOKEN }}
          image: rhacs-eng/${{ matrix.image }}:${{ inputs.version }}
          limit: 1800

  deploy:
    runs-on: ubuntu-latest
    needs: [wait-for-images]
    steps:
      - name: Show inputs
        run: |
          echo "Environment: ${{ inputs.environment }}"
          echo "Version: ${{ inputs.version }}"

      - name: Check out code
        uses: actions/checkout@v7
        with:
          fetch-depth: 0
          ref: ${{ inputs.version }}

      - name: Install OpenShift CLI
        uses: redhat-actions/openshift-tools-installer@v1
        with:
          oc: latest
          
      - name: Authenticate to OpenShift
        uses: redhat-actions/oc-login@v1
        with:
          openshift_server_url: ${{ secrets.OCP4_SERVER_URL }}
          openshift_token: ${{ secrets.OCP4_TOKEN }}
          insecure_skip_tls_verify: false

      - name: Authenticate to GCP (for secrets)
        uses: google-github-actions/auth@v2
        with:
          credentials_json: ${{ secrets.GCP_SA_KEY }}

      - name: Deploy to ${{ inputs.environment }}
        run: |
          ENVIRONMENT=${{ inputs.environment }} \
          VERSION=${{ inputs.version }} \
          ./scripts/deploy/ocp4-helm.sh deploy ${{ inputs.version }} ${{ inputs.environment }}

      - name: Notify infra channel about new version
        uses: slackapi/slack-github-action@v3.0.5
        with:
          method: chat.postMessage
          token: ${{ secrets.SLACK_BOT_TOKEN }}
          payload: |
            channel: "CVANK5K5W"
            text: "Infra OCP 4 (${{ inputs.environment }}) was updated."
            blocks:
              - type: "section"
                text:
                  type: "mrkdwn"
                  text: ":ship::tada:*Infra OCP 4 (${{ inputs.environment }}) was updated to <${{ github.server_url }}/${{ github.repository }}/releases/tag/${{ inputs.version }}|${{ inputs.version }}>."
```

## Prerequisites

Before implementation, ensure:

1. **External Secrets Operator installed on OCP 4 cluster:**
```bash
oc apply -f https://raw.githubusercontent.com/external-secrets/external-secrets/main/deploy/crds/bundle.yaml
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets-system --create-namespace
```

2. **GCP Service Account for External Secrets:**
Create service account with Secret Manager read access and configure in OCP 4 as a secret in `external-secrets-system` namespace.

3. **GitHub Secrets configured:**
- `OCP4_SERVER_URL` - OpenShift cluster API URL
- `OCP4_TOKEN` - Service account token with cluster-admin or appropriate RBAC
- `GCP_SA_KEY` - GCP service account key for secret access
- `QUAY_RHACS_ENG_BEARER_TOKEN` - Existing Quay token
- `SLACK_BOT_TOKEN` - Existing Slack token

4. **OCP 4 cluster requirements:**
- Cluster admin access
- External Secrets Operator CRDs installed
- Sufficient quotas for namespace resources
- StorageClass available (e.g., gp3-csi, gp2, or OCS)
- Network egress to GCP Secret Manager and GCS

## Testing & Validation

### 1. Local Template Validation
```bash
ENVIRONMENT=development VERSION=latest make ocp4-helm-template > /tmp/rendered.yaml
# Review rendered manifests
oc apply --dry-run=client -f /tmp/rendered.yaml
```

### 2. Development Deployment
```bash
oc login <ocp4-cluster-url>
ENVIRONMENT=development VERSION=<latest-tag> make ocp4-helm-deploy
```

### 3. Functional Validation
- Verify route is accessible: `curl -k https://dev.infra.stackrox.com/healthz`
- Test infractl connectivity: `infractl -e dev.infra.stackrox.com:443 whoami`
- Create test cluster: `infractl -e dev.infra.stackrox.com:443 cluster create --flavor demo`
- Check Argo Workflows: `oc port-forward -n argo svc/infra-server-argo-workflows-server 2746`
- Verify monitoring: Access Prometheus/Grafana dashboards
- Test all cloud provider workflows (GKE, AWS, Azure, etc.)

### 4. Production Deployment
```bash
oc login <ocp4-prod-cluster-url>
ENVIRONMENT=production VERSION=<release-tag> make ocp4-helm-deploy
```

### 5. DNS Cutover
Update DNS records to point to OCP 4 routes:
- `infra.rox.systems` → OpenShift Route
- `dev.infra.rox.systems` → OpenShift Route

## Rollback Strategy

If issues occur after deployment:

**Immediate (< 1 hour):**
```bash
helm rollback infra-server -n infra
```

**DNS revert (if cutover completed):**
Revert DNS records to GKE ingress IPs

**Complete rollback:**
1. Restore DNS to GKE
2. Verify GKE deployment health
3. Document issues for remediation

## Success Criteria

- [ ] All Helm templates render without errors
- [ ] Pods start successfully with OCP security contexts
- [ ] Routes accessible with valid TLS certificates
- [ ] External Secrets sync from GCP Secret Manager
- [ ] Argo Workflows execute successfully with GCS artifacts
- [ ] Monitoring metrics collected by Prometheus
- [ ] All infractl commands work (cluster create, list, delete)
- [ ] All cloud provider flavors functional (GKE, AWS, Azure, ARO, ROSA, etc.)
- [ ] Performance meets or exceeds GKE baseline
- [ ] Documentation updated (DEPLOYMENT.md, README.md)
