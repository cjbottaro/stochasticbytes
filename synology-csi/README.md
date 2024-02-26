# Synology CSI Chart for Kubernetes

https://github.com/christian-schlichtherle/synology-csi-chart.git

Edit `values.yaml` to have the Diskstation IP address, username, and password.

Then run:
```
helm repo add synology-csi-chart https://christian-schlichtherle.github.io/synology-csi-chart
helm install synology-csi synology-csi-chart/synology-csi --namespace kube-system
```