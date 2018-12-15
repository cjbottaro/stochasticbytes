{{- define "vpn-volumes" }}
- name: vpn-config
  configMap:
    name: openvpn
- name: vpn-secret
  secret:
    secretName: openvpn
- name: dev-net
  emptyDir: {}
{{- end -}}

{{- define "vpn-container" }}
- name: openvpn
  image: cjbottaro/openvpn:latest
  args:
    - --config
    - /etc/vpn.ovpn
  securityContext:
    capabilities:
      add:
        - NET_ADMIN
  volumeMounts:
    - name: vpn-config
      mountPath: /etc/vpn.ovpn
      subPath: vpn.ovpn
    - name: vpn-secret
      mountPath: /etc/vpn.auth
      subPath: vpn.auth
    - name: dev-net
      mountPath: /dev/net
{{- end -}}
