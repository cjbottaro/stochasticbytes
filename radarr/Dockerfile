FROM frolvlad/alpine-mono

ARG RADARR_VERSION

RUN apk add --no-cache libmediainfo
ADD https://github.com/Radarr/Radarr/releases/download/v$RADARR_VERSION/Radarr.develop.$RADARR_VERSION.linux.tar.gz /Radarr.tar.gz
RUN tar xzf Radarr.tar.gz
WORKDIR Radarr

ENTRYPOINT ["mono", "Radarr.exe", "-nobrowser"]
