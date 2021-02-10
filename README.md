# Kubernetes Webhook

## Setup

```bash
kind create cluster --config=cluster.yaml
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.1.0/cert-manager.yaml
kubectl create namespace apps
kubectl config set-context --current --namespace=apps
```

```bash
kubectl apply -f test-app.yaml
```

## Deployment

```bash
make docker-build kind-load deploy IMG=padok.fr/webhook:v1.0.0-$(date +%s)
```

## Usage

```bash
kubectl label namespace apps padok.fr/inject-tolerations=true
kubectl rollout restart deployment nginx-deployment
```
