FROM certbot/certbot
RUN pip install certbot-dns-dnsimple
ADD https://storage.googleapis.com/kubernetes-release/release/v1.23.5/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chmod 755 /usr/local/bin/kubectl