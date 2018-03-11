set -e

certbot $@ --cert-name mycert

cp /etc/letsencrypt/live/mycert/privkey.pem /secrets/privkey.pem
cp /etc/letsencrypt/live/mycert/fullchain.pem /secrets/fullchain.pem
