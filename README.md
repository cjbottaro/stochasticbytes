# stochasticbytes

This is for managing my VPS nodes, Docker images, Docker swarm, and personal websites.

## Setup Raspberry Pi

Prepare Rasberry Pi.

### Install various packages

```sh
apt-get update
apt-get install -y vim
apt-get upgrade -y
reboot
```

### Disable swap

```sh
dphys-swapfile swapoff
dphys-swapfile uninstall
systemctl disable dphys-swapfile
```

### Disable Wifi and Bluetooth

```sh
cat <<EOF >> /boot/config.txt
dtoverlay=disable-wifi
dtoverlay=disable-bt
EOF
reboot
```

## Kubernetes cluster

Install master node, worker node, helm, etc.

### Install Docker

https://kubernetes.io/docs/setup/production-environment/container-runtimes/#docker

Needs to be modified a little bit to work with Raspbian (Buster).

```sh
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -

cat <<EOF >/etc/apt/sources.list.d/docker.list
deb https://download.docker.com/linux/raspbian $(lsb_release -cs) stable
EOF

apt-get update && apt-get install --no-install-recommends -y docker-ce

cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2"
}
EOF

systemctl daemon-reload
systemctl restart docker
```

### Install kubeadm

https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/

```sh
apt-get update && apt-get install -y apt-transport-https curl

curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

cat <<EOF > /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF

apt-get update
apt-get install -y kubelet kubeadm kubectl
apt-mark hold kubelet kubeadm kubectl
```

### Enable legacy iptables

Apparently kubeadm doesn't work with nftables which is what Raspbian Buster uses.

https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#ensure-iptables-tooling-does-not-use-the-nftables-backend

```sh
update-alternatives --set iptables /usr/sbin/iptables-legacy
update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
update-alternatives --set arptables /usr/sbin/arptables-legacy
update-alternatives --set ebtables /usr/sbin/ebtables-legacy
```

### Run kubeadm init

```sh
kubeadm init --pod-network-cidr=100.64.0.0/16
```

Then update your `secrets.yaml` with the output of that command.
```yaml
kubeadm:
  join_cmd: <blah>
```

## Install pod networking

```
kubectl apply -f https://raw.githubusercontent.com/cloudnativelabs/kube-router/master/daemonset/kubeadm-kuberouter-all-features-hostport.yaml
```

Unfortunately, the kube-router manifests only support amd64, so we gotta do a little tweaking.

Edit the daemonset.
```
kubectl edit ds -n kube-system kube-router
```

And patch.
```yaml
spec:
  template:
    spec:
      nodeSelector:
        beta.kubernetes.io/arch: amd64
```
That will keep the pods off arm machines (like Raspberry Pi).

Now we need to get the right ones on our arm machines.
```
kubectl get ds -n kube-system kube-router -o yaml > kube-router-arm.yaml
```

Edit that file and patch.
```yaml
metadata:
  name: kube-router-arm
spec:
  template:
    spec:
      nodeSelector:
        beta.kubernetes.io/arch: arm
      containers:
        - image: xjjo/kube-router:arm-v0.3.2
```

Apply it.
```sh
kubectl apply -f kube-router-arm.yaml
```

Don't forget to get rid of `kube-proxy`.
```sh
kubectl -n kube-system delete ds kube-proxy
```

You should now have a working Kubernetes cluster.

## Install tiller (Helm)

```sh
kubectl apply -f tiller-rbac.yaml

# If using Helm > 2.14.3
helm init --service-account tiller

# Workaround for Helm <= 2.14.3 and Kubernetes 1.16.x
# https://github.com/helm/helm/issues/6374#issuecomment-533427268
helm init --service-account tiller --override spec.selector.matchLabels.'name'='tiller',spec.selector.matchLabels.'app'='helm' --output yaml | sed 's@apiVersion: extensions/v1beta1@apiVersion: apps/v1@' | kubectl apply -f -

# Use arm build of tiller, set image: jessestuart/tiller:v2.14.3
kubectl -n kube-system edit deployment tiller-deploy
```