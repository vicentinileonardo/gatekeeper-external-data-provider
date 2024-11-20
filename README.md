# External Data Provider

This repo was created in the context of my master's thesis. 
Currently in standby since limitations were found in the Gatekeeper project.

## Setup

Build and deploy the external data provider.

```bash
git clone https://github.com/open-policy-agent/gatekeeper-external-data-provider.git
cd gatekeeper-external-data-provider

# Generate a self-signed certificate for the external data provider
# see: https://open-policy-agent.github.io/gatekeeper/website/docs/externaldata/#how-to-generate-a-self-signed-ca-and-a-keypair-for-the-external-data-provider
./scripts/generate-tls-cert.sh

```

The following step assume the use of `minikube` as the Kubernetes cluster. If you are using a different cluster, you may need to adjust the step related to the docker daemon.

```bash
# Point your shell to minikube's docker-daemon
eval $(minikube docker-env)
docker ps
docker images

# Build the image
docker build -t openpolicyagent/gatekeeper-external-data-provider:dev .

# Deploy the external data provider (client and server auth enabled)
helm install external-data-provider charts/external-data-provider \
    --set provider.tls.caBundle="$(cat certs/ca.crt | base64 | tr -d '\n\r')" \
    --namespace "${NAMESPACE:-gatekeeper-system}"

# Verify the external data provider is running
kubectl get pods -n "${NAMESPACE:-gatekeeper-system}"
kubectl get deployments -n "${NAMESPACE:-gatekeeper-system}"
kubectl get services -n "${NAMESPACE:-gatekeeper-system}"

kubectl get providers.externaldata.gatekeeper.sh external-data-provider -n "${NAMESPACE:-gatekeeper-system}"

# Install Assign mutation
kubectl apply -f mutation/assign-scheduling.yaml

# Verify the mutation is running
kubectl get assign.mutations.gatekeeper.sh assign-scheduling -n "${NAMESPACE:-gatekeeper-system}"

# Get pod logs 
kubectl logs $(kubectl get pods -n "${NAMESPACE:-gatekeeper-system}" -o jsonpath='{.items[0].metadata.name}') -n "${NAMESPACE:-gatekeeper-system}"
```

## Uninstalling

```bash
kubectl delete -f mutation/

helm uninstall external-data-provider --namespace "${NAMESPACE:-gatekeeper-system}"

docker rmi openpolicyagent/gatekeeper-external-data-provider:dev
```
