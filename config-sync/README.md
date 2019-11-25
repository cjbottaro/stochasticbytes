# config-sync

Plex, Emby, Radarr, and Sonarr all use SQLite which really doesn't like its files served
from NAS. The workaround is to have a cronjob that rsync's each application's config dir
to the NAS. Then each application should have an initContainer that checks the hostPath
of the config dir and if it doesn't exist, rsync it from the NAS.