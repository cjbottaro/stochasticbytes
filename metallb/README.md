# Add the metallb repo

```
helm repo add metallb https://metallb.github.io/metallb
helm repo update
```

# Install/Upgrade


## Skylake
```
helm install metallb metallb/metallb -n kube-system -f metallb/values.yaml
```

See: https://metallb.universe.tf/installation/#upgrade

```
helm upgrade metallb metallb/metallb -f metallb/values.yaml
```
