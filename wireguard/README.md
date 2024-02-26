## Wireguard

This is much easier to setup than OpenVPN. It also doesn't require any process to
be running, so it can all be done via an initContainer.

## Generate public and private keys

```
apk add wireguard-tools
wg genkey | tee keys | wg pubkey >> keys && cat keys && rm -f keys

yERY8iuE3utvGfcOSRkC5sQXmA/hneo+neQ9QxJ260M= # private
5UwHIMyTnBznf+pGdNmnWqs6SIDCnc/huIvOGtykHTA= # public
```

## Get an allocated IP address

You then upload your _public key_ to your VPN provider and they will
give you an IP address in return. In this example, we get `172.23.248.201`.

## Config

The config is saved to something like `/etc/wireguard/vpn`.

```
[Interface]
PrivateKey = yERY8iuE3utvGfcOSRkC5sQXmA/hneo+neQ9QxJ260M=
Address = 172.23.248.201/32
DNS = 172.16.0.1

[Peer]
PublicKey = <endpoint-public-key>
Endpoint = us-tx1.wg.ivpn.net:2049
AllowedIPs = 0.0.0.0/0
```

`Address` is given when you upload your public key to your VPN provider.

`DNS` is given by your VPN provider, _but leave it blank_. As in, do not
have a `DNS` section at all. We want to use Koob's internal DNS so that
services resolve.

`PublicKey` is the _endpoint's_ public key, not _your_ public key. It is
provided by your VPN provider.

`Endpoint` is provided by your VPN provider.

## AllowedIPs

This is where the magic happens. Really we want to do this:
```
AllowedIPs = 0.0.0.0/0
DisallowedIPs = <endpoint-ip-address>/32, 10.0.0.0/8, 100.64.1.0/24, 192.168.1.0/24,
```

`<endpoing-ip-address>` would be the IP address of `us-tx1.wg.ivpn.net` in the above
config example.

`10.0.0.0/8` are Koob service IPs.

`100.64.1.0/24` are Koob pod IPs.

`192.168.1.0/24` are LAN IPs (like Synology Diskstation).

Note these IPs can be different depending on how your LAN or Koob is setup.

Unfortunately, `DisallowedIPs` is not a thing, so we have to use a CIDR calculator
to figure out what to set `AllowedIPs` to.

https://www.procustodibus.com/blog/2021/03/wireguard-allowedips-calculator/

**IMPORTANT** Make sure you put the endpoint's ip address into the calculator.

## Starting

```
wg-quick up vpn
```

This affects the network of _all containers_ in a pod. Thus it can be used
in a sidecar or initContainer. The latter is preferred since there is no
need to keep a process up and running.

It's fully isolated; it won't affect the host node at all, or any other pods
in the cluster.