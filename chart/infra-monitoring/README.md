# Monitoring

## Requirements

The infra cluster must have sufficient resources.
The development cluster was scaled to 2 nodes per zone for that.

## Setup

Deploying monitoring stack:

```bash
helm upgrade prometheus-stack chart/infra-monitoring \
    --install \
    --namespace monitoring \
    --create-namespace \
    --version 8.27.0 \
    --values chart/infra-monitoring/values.yaml \
    --wait
```

Exposing metrics from Argo Workflow Controller:

```bash
kubectl apply -f monitoring/argo.yaml
```

Creating alertmanager config with the webhook URL from the Bitwarden Secret `Infra Service - Slack App Webhook`:

```bash
kubectl create secret generic alertmanager-slack-webhook \
    --namespace monitoring \
    --from-literal webhookURL=https://...
kubectl apply -f monitoring/alertmanager.yaml
```
