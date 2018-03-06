# stochasticbytes

This is for managing my VPS nodes, Docker images, Docker swarm, and personal websites.

## Creating an Alpine image in DigitalOcean

You like Alpine. As of this writing, DigitalOcean does not support Alpine natively.
Luckily, someone made a script to install Alpine on a Debian Droplet. You can then
save the Alpine install as an image and create Droplets from it.

Launch a Debian 9.3 Droplet, then run this:
```
wget -q https://github.com/bontibon/digitalocean-alpine/raw/master/digitalocean-alpine.sh
chmod 755 digitalocean-alpine.sh
./digitalocean-alpine.sh --rebuild
```

After it reboots, power it down, and create an image from it.

## Initializing the Docker swarm

Bring up any number of nodes in DigitalOcean using your Alpine image.

Create a file at `ansible/inventory.ini` like:
```
[main_manager]
10.0.0.1

[managers]
10.0.0.1
10.0.0.2

[workers]
10.0.0.3
10.0.0.4
10.0.0.5
```

Now run Ansible. Since it's all Dockerized, it should just work...
```
docker-compose run --rm ansible swarm_init.yml
```

## Adding nodes to the swarm

Since you took great care in making the dang Ansible playbook idempotent, adding
nodes should be as easy as modifying `ansible/inventory.ini` and rerunning:
```
docker-compose run --rm ansible swarm_init.yml
```
