# Exported via Kubernetes secret
# TOKEN="your-oauth-token"  # The API v2 OAuth token

ACCOUNT_ID="7481"      # Replace with your account ID
ZONE_ID="cjbotta.ro"   # The zone ID is the name of the zone (or domain)
RECORD_ID="13635713"   # Replace with the Record ID
IP=`curl --ipv4 -s http://icanhazip.com/`

curl -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -H "Accept: application/json" \
     -X "PATCH" \
     -i "https://api.dnsimple.com/v2/$ACCOUNT_ID/zones/$ZONE_ID/records/$RECORD_ID" \
     -d "{\"content\":\"$IP\"}"
