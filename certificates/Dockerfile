FROM certbot/certbot
RUN pip install certbot-dns-dnsimple
ADD shim.sh .
ENTRYPOINT ["/bin/sh", "shim.sh"]
