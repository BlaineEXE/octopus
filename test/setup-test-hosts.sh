#!/usr/bin/env bash
set -Eeuo pipefail

# Run some test hosts, and generate a node list with group "one" the first ip, "rest" the
# rest of the ips and "all" set to all of the ips
one=""
rest=""
for i in $(seq 1 $NUM_HOSTS); do
  host=$HOST_BASENAME-$i
  echo "running test host $host"
  docker run --rm --name "$host" --hostname "$host" --detach $IMAGE_TAG
  ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$host")
  echo "ip: $ip"
  if [ $i = 1 ] ; then
    one="$ip"
  else
    rest="$(printf '%s\n%s' "$rest" "$ip")"
  fi
done

cat << EOF > "$GROUPFILE"
#!/usr/bin/env bash

export one='$one'
export rest='$rest'
export empty=''

EOF
echo 'export all="$one $rest"' >> "$GROUPFILE"
echo '' >> "$GROUPFILE"

echo "  "$GROUPFILE" file:"
cat "$GROUPFILE"
