# Ops Notes

Context for future maintainers picking up ingress and registry work.

## Current state (Dec 2025)
- k3s default Traefik is the ingress controller; per-app Ingress + Service lives in each chart (plex, jellyfin, radarr, sonarr, sabnzbd done).
- The old standalone Nginx reverse-proxy Helm release is **not installed**, and the `nginx/` chart has been removed from the repo.
- MetalLB: `values.yaml` still has `ingress.ip_address` 192.168.1.78 (Traefik); old `nginx.ip_address` references may linger in docs/DNS. Ensure `*.cjbotta.ro` points to the Traefik IP.
- Registry (`registry/` chart) currently has only a Deployment (no Service/Ingress), so it is not exposed externally.

## If you need the registry again
- Prefer exposing it via Traefik instead of resurrecting the Nginx proxy.
- Add a Service on port 5000 with selector `app: registry`, and an Ingress for `registry.cjbotta.ro` using Traefik entrypoints `web,websecure`.
- Include middleware/annotations for basic auth and large pushes, e.g.:
  - Basic auth: define a Traefik middleware secret and reference it with `traefik.ingress.kubernetes.io/router.middlewares: default-registry-auth@kubernetescrd`.
  - Allow big pushes: `traefik.ingress.kubernetes.io/buffering: "true"` and `traefik.ingress.kubernetes.io/buffering.maxrequestbodybytes: "0"` (or equivalent config that matches your Traefik version).
- Remember to keep `client_max_body_size 0` equivalent for registry only, not the whole ingress.
- After adding Service/Ingress, update DNS for `registry.cjbotta.ro` to point at the Traefik IP.

## Helm/Kubeconfig
- Use `KUBECONFIG=./kube-config.dec.yaml` with `bin/upgrade <chart>` to deploy changes.
- Helm releases currently installed: jellyfin, plex, radarr, sonarr, sabnzbd, wireguard (see `helm list`). There is no `nginx` release.
