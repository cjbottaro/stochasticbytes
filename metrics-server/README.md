This is so `kubectl top pods` works.

```
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm upgrade --install metrics-server metrics-server/metrics-server --namespace kube-system
kubectl edit deployment metrics-server -n kube-system
```

Add the following argument to the container args:
```
--kubelet-insecure-tls
```