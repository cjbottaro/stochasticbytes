ARG ALPINE_VERSION
FROM alpine:$ALPINE_VERSION AS build
ARG ANSIBLE_VERSION
RUN apk add --no-cache python2-dev py2-pip build-base libffi-dev openssl-dev
RUN pip install ansible==$ANSIBLE_VERSION

ARG ALPINE_VERSION
FROM alpine:$ALPINE_VERSION
RUN apk add --no-cache python2 libffi openssl openssh-client
COPY --from=build /usr/lib/python2.7/site-packages /usr/lib/python2.7/site-packages
COPY --from=build /usr/bin/ansible* /usr/bin/
ENTRYPOINT ["ansible"]
