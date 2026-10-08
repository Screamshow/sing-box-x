#!/bin/sh
# Private configs remain on the test VM; output uses numeric node IDs only.
set -eu
umask 077
binary=$1
configs=$2
results=$3
mkdir -p "$results"
probe() (
    id=$1 port=$2 transport=$3 fingerprint=$4
    process=
    trap '[ -z "$process" ] || { kill "$process" 2>/dev/null || true; wait "$process" 2>/dev/null || true; }' EXIT INT TERM
    if ! "$binary" check -c "$configs/node-$id.json" > "$results/node-$id.check.log" 2>&1; then
        printf '%s\t%s\t%s\tcheck-failed\n' "$id" "$transport" "$fingerprint" > "$results/node-$id.tsv"
        exit 0
    fi
    "$binary" run -c "$configs/node-$id.json" > "$results/node-$id.runtime.log" 2>&1 &
    process=$!
    ready=0
    for attempt in 1 2 3 4 5; do
        if netstat -lnt 2>/dev/null | grep -q "127.0.0.1:$port "; then ready=1; break; fi
        sleep 1
    done
    if [ "$ready" != 1 ] || ! kill -0 "$process" 2>/dev/null; then
        printf '%s\t%s\t%s\trun-failed\n' "$id" "$transport" "$fingerprint" > "$results/node-$id.tsv"
        exit 0
    fi
    first_status=0 second_status=0
    first=$(curl -sS --connect-timeout 5 --max-time 10 --proxy "http://127.0.0.1:$port" -o /dev/null -w '%{http_code}' https://www.gstatic.com/generate_204 2> "$results/node-$id.curl.log") || first_status=$?
    second=$(curl -sS --connect-timeout 5 --max-time 10 --proxy "http://127.0.0.1:$port" -o "$results/node-$id.response" -w '%{http_code}' https://example.com 2>> "$results/node-$id.curl.log") || second_status=$?
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$id" "$transport" "$fingerprint" "$first_status" "$first" "$second_status" "$second" > "$results/node-$id.tsv"
)
batch=0
while read -r id port transport fingerprint; do
    probe "$id" "$port" "$transport" "$fingerprint" &
    batch=$((batch+1))
    if [ "$batch" -eq 4 ]; then wait; batch=0; fi
done < "$configs/nodes.tsv"
wait
cat "$results"/node-*.tsv | sort -n > "$results/results.tsv"
awk -F '\t' '{total++; if($4==0 && $5==204 && $6==0 && $7==200) passed++; else failed++} END {printf "nodes=%d passed=%d failed=%d\n", total, passed, failed}' "$results/results.tsv"
