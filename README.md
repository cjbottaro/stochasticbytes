# stochasticbytes

Home Kubernetes cluster.

## Alpine on Raspberry Pi

These instructions are adapted from this wiki:
[https://wiki.alpinelinux.org/wiki/Raspberry_Pi_-_Headless_Installation](https://wiki.alpinelinux.org/wiki/Raspberry_Pi_-_Headless_Installation)

We don't need multiple partitions on the SD card.
```sh
# Find the device
diskutil list

# Erase the SD card
sudo diskutil eraseDisk FAT32 ALPINE MBRFormat /dev/diskX

# Copy files
cp xf ~/Downloads/alpine-rpi-3.14.0-aarch64.tar.gz -C /Volumes/ALPINE
cp ~/Downloads/headless.apkovl.tar.gz /Volumes/ALPINE

# Eject
sudo diskutil umount /Volumes/ALPINE

# SSH with no password
ssh root@clawhammer

# Setup Alpine
setup-alpine

# This is optional; the headless overlay somehow makes this work already
cat <<EOF >> /etc/ssh/sshd_config
PermitRootLogin yes
EOF

# Restart SSH
service sshd restart
```

The Google Fiber router remembers devices by MAC address, so you can still ssh by hostname.
If that doesn't work, you gotta find the assigned IP address in the list of devices.

Alpine runs in-memory by default, so when `setup-alpine` asks, you want to install `sys` to
the SD card.

## Alpine on Media Server

The media server is a normal x86_64 machine with a normal hard drive and everything. We put
Alpine on a bootable USB drive and install from there.

Unfortunately, there is no headless overlay, so you gotta plug in a monitor and keyboard.

```sh
# Find device
diskutil list

# Write image to device
sudo dd if=~/Downloads/alpine-standard-3.14.0-x86_64.iso of=/dev/diskX bs=4M oflag=sync status=progress
```

Once on the machine...

```sh
# Setup Alpine
setup-alpine

# This is required
cat <<EOF >> /etc/ssh/sshd_config
PermitRootLogin yes
EOF

# Restart SSH
service sshd restart
```

The Google Fiber router remembers devices by MAC address, so you can still ssh by hostname.
If that doesn't work, you gotta find the assigned IP address in the list of devices.

Alpine runs in-memory by default, so when `setup-alpine` asks, you want to install `sys` to
the hard drive.

## Kubernetes

```sh
setup-alpine
```

This setups the rest of the environment (cgroup stuff, Docker, Kubernetes, other packages).
```sh
wget -qO- https://bit.ly/3jleHNB | sh
```

Both need to do.

```sh
# Disable swap
swapoff -a
vim /etc/fstab

# Uncomment edge/community and edge/testing
vim /etc/apk/repositories
apk update

# Install Kubernetes, kubeadm, etc
apk add open-iscsi eudev blkid xfsprogs kubernetes kubeadm kubelet cni-plugins docker
rc-update add docker default
rc-update add kubelet default
```

Init a master.

```sh
# Init cluster
kubeadm init --apiserver-cert-extra-sans skyranger.cjbotta.ro --pod-network-cidr=100.64.0.0/16

# Setup networking
kubectl apply -f https://raw.githubusercontent.com/cloudnativelabs/kube-router/master/daemonset/kubeadm-kuberouter.yaml
```

It will output the location of your kubeconfig (`/etc/kubernetes/admin.conf`) and
a join cluster command. Need to write that down somewhere (1Password).

# Media Server

Running a media server in Kubernetes is challenging.

## VPN and routing issues

Even though it's possible to run VPN connections as sidecar containers, it screws with routing somehow: The container that is using the VPN won't be reachable by containers running on a different node.

The workaround is to put all containers with a VPN sidecar _and all the containers that need to talk to them_ on the same node.

```yaml
nodeSelector:
  kubernetes.io/hostname: pineview
```

## SQLite and NAS

Just don't do it (as in, don't serve SQLite databases from your NAS). You'll run into all sorts of locking issues and your application will basically be broken.

Use iSCSI instead (see below).

## iSCSI

Need to install on all worker nodes (and maybe reboot):
```
apk add open-iscsi eudev blkid xfsprogs
```

`eudev` is needed so that `/dev/disks/by-path/` works.

The default `blkid` provided by Alpine is not compatible with the Kubernetes iSCSI
provisioner, thus we have to install the `blkid` package which give us the "standard"
Linux version.

`xfsprogs` is if you want to format the iSCSI devices using XFS.

Then you can go into DSM and create iSCSI targets and luns.

In a Kubernetes resource, the `targetPortal` _must_ be an ip address for some reason...
```yaml
volumes:
  - name: app-data
    iscsi:
      targetPortal: 192.168.1.113
      iqn: iqn.2000-01.com.synology:diskstation.Target-1.77d659b9be
      lun: 1
      fsType: ext4
```

Some useful iSCSI commands...

Discover all targets.
```
iscsiadm --m discovery -p diskstation -t sendtargets
```

List all sessions. An active session should have a corresponding device in `/dev/disks/by-path/`.
```
iscsiadm -m session
```

Manuall log into a session (useful to testing DSM targets).
```
iscsiadm -m node -p diskstation -T iqn.2000-01.com.synology:diskstation.Target-1.77d659b9be --login
```

Log out of a session.
```
iscsiadm -m node -p diskstation -u -T iqn.2000-01.com.synology:diskstation.Target-1.77d659b9be
```