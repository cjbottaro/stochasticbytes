# stochasticbytes

This is for managing my VPS nodes, Docker images, Docker swarm, and personal websites.

## Setup Raspberry Pi

```sh
sudo bash

# Set hostname and timezone
raspi-config

# Install stuff and upgrade
apt-get update
apt-get install -y vim
apt-get upgrade -y

# Disable swap
dphys-swapfile swapoff
dphys-swapfile uninstall
systemctl disable dphys-swapfile

# Disable wifi and bluetooth
cat <<EOF >> /boot/config.txt
dtoverlay=disable-wifi
dtoverlay=disable-bt
EOF

# Setup ssh
mkdir .ssh
cat <<EOF >> .ssh/authorized_keys
ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA0vALsY+6CaDuhjZ2X/bXOnNReEfzvQsdAs6Iex0Hg/s+I4W3ydLk99turzYgic1jA4eshXhHPaY5Oh1Cs//6cmfdoB45u6uoqGdCzO/lVekYTE8wmoq4c7bUWv7nJT7VWH7xurGRIVlkXjy65Z9Jo//JBvnpYnWps79E9pHtbMiEhV5oWoa105GAyb3/RGJcnv0MaXkYpwEKiPyz9iPVhwDFzBfcfr7NPneFcWtvs9TimcrjWaUXGvEL+wDwNEyBkj5WJRMadl/PeKfGCESNAP00IKYO81MtX9eiGgA0mvOOm6cBWVcNezDxZOcrPuYTOGj2skz3s1vWDEgDXpcDiQ== cjbottaro
EOF
chmod 700 .ssh
chmod 600 .ssh/authorized_keys

# Reboot
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
deb https://download.docker.com/linux/$(lsb_release -is | tr "[:upper:]" "[:lower:]") $(lsb_release -cs) stable
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

```sh
kubectl -n kube-system apply -f kube-router.yaml
kubectl -n kube-system delete ds kube-proxy
```

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

# Media Server

Running a media server in Kubernetes is challenging.

## VPN and routing issues

Even though it's possible to run VPN connections as sidecar containers, it screws with routing somehow: The container that is using the VPN won't be reachable by containers running on a different node.

The workaround is to put all containers with a VPN sidecar _and all the containers that need to talk to them_ on the same node.

```yaml
nodeSelector:
  kubernetes.io/hostname: vpn-node-name
```

## SQLite and NAS

Just don't do it (as in, don't serve SQLite databases from your NAS). You'll run into all sorts of locking issues and your application will basically be broken.

Workaround is to run a cronjob that rsyncs application's config dir to NAS, then have initContainers that check for the existance of the config dir on hostPath volumes and if they don't exist, rsync it from NAS.

WARNING: Make sure the cronjobs are not running when the initContainers are rsync'ing from NAS to hostPath.

## 