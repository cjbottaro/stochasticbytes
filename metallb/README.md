# Install

https://metallb.universe.tf/installation

```
helm repo add metallb https://metallb.github.io/metallb
helm repo update
helm install metallb metallb/metallb -n kube-system
kubectl apply -f address_pool.yaml
kubectl apply -f advertisement.yaml
```