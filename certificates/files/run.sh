certbot certonly \
--non-interactive \
--agree-tos \
-m {{ .Values.email }} \
--dns-dnsimple \
--dns-dnsimple-credentials /etc/dnsimple.ini \
--dns-dnsimple-propagation-seconds 60 \
-d {{ .Values.domains | join "," }}

# Update secret with kubectl https://stackoverflow.com/a/45881259
kubectl create secret generic certificates \
--from-file=/etc/letsencrypt/live/{{ index .Values.domains 0 }}/fullchain.pem \
--from-file=/etc/letsencrypt/live/{{ index .Values.domains 0 }}/privkey.pem \
--dry-run=client \
-o yaml \
| kubectl apply -f -

# kubectl patch deployment nginx --patch "{
#   \"spec\": {
#     \"template\": {
#       \"metadata\": {
#         \"annotations\": {
#           \"deploy/timestamp\": \"$(date)\"
#         }
#       }
#     }
#   }
# }"