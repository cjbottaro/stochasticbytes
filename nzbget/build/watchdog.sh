echo "[watchdog] starting"
while true; do
  if ! ifconfig tun0 > /dev/null 2>&1; then
    echo "[watchdog] tunnel is down, terminating nzbget"
    kill $(pgrep nzbget)
  fi
  sleep 1
done
