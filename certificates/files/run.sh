certbot certonly \
--non-interactive \
--agree-tos \
-m {{ .Values.account }} \
--dns-dnsimple \
--dns-dnsimple-credentials /etc/dnsimple.ini \
-d {{ .Values.domains | join "," }}

# Update secret with kubectl https://stackoverflow.com/a/45881259
kubectl create secret generic certificates \
--from-file=/etc/letsencrypt/live/cjbotta.ro/cert.pem \
--from-file=/etc/letsencrypt/live/cjbotta.ro/chain.pem \
--from-file=/etc/letsencrypt/live/cjbotta.ro/fullchain.pem \
--from-file=/etc/letsencrypt/live/cjbotta.ro/privkey.pem \
--dry-run \
-o yaml \
| kubectl apply -f -

kubectl patch deployment nginx --patch "{
  \"spec\": {
    \"template\": {
      \"metadata\": {
        \"annotations\": {
          \"deploy/timestamp\": \"$(date)\"
        }
      }
    }
  }
}"