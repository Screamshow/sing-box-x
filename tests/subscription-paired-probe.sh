#!/bin/sh
# Compare the same node at the same time, without parallel batches of other nodes.
set -eu
umask 077
baseline=$1 candidate=$2 old_configs=$3 new_configs=$4 results=$5
mkdir -p "$results"
old_pid= new_pid= old_request= new_request=
cleanup() {
    for pid in "$old_request" "$new_request" "$old_pid" "$new_pid"; do
        [ -z "$pid" ] || { kill "$pid" 2>/dev/null || true; wait "$pid" 2>/dev/null || true; }
    done
    old_pid= new_pid= old_request= new_request=
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
while read -r id old_port transport fingerprint; do
    new_port=$(awk -F '\t' -v id="$id" '$1==id {print $2}' "$new_configs/nodes.tsv")
    "$baseline" run -c "$old_configs/node-$id.json" > "$results/$id.old.log" 2>&1 & old_pid=$!
    "$candidate" run -c "$new_configs/node-$id.json" > "$results/$id.new.log" 2>&1 & new_pid=$!
    for attempt in 1 2 3 4 5; do
        if netstat -lnt 2>/dev/null | grep -q "127.0.0.1:$old_port " && netstat -lnt 2>/dev/null | grep -q "127.0.0.1:$new_port "; then break; fi
        sleep 1
    done
    kill -0 "$old_pid" && kill -0 "$new_pid"
    (code=0; http=$(curl -sS --max-time 6 --proxy "http://127.0.0.1:$old_port" -o /dev/null -w '%{http_code}' https://www.gstatic.com/generate_204 2> "$results/$id.old.curl.log") || code=$?; printf '%s\t%s\n' "$code" "$http" > "$results/$id.old.tsv") & old_request=$!
    (code=0; http=$(curl -sS --max-time 6 --proxy "http://127.0.0.1:$new_port" -o /dev/null -w '%{http_code}' https://www.gstatic.com/generate_204 2> "$results/$id.new.curl.log") || code=$?; printf '%s\t%s\n' "$code" "$http" > "$results/$id.new.tsv") & new_request=$!
    wait "$old_request"; wait "$new_request"; old_request= new_request=
    printf '%s\t%s\t%s\t%s\t%s\n' "$id" "$transport" "$fingerprint" "$(cat "$results/$id.old.tsv")" "$(cat "$results/$id.new.tsv")" > "$results/$id.tsv"
    cleanup
    if [ $((id%10)) -eq 0 ]; then echo "paired_nodes_completed=$id"; fi
done < "$old_configs/nodes.tsv"
cat "$results"/[0-9]*.tsv | awk -F '\t' 'NF==7' | sort -n > "$results/results.tsv"
awk -F '\t' '{total++; old=($4==0 && $5==204); new=($6==0 && $7==204); if(old&&new) both++; else if(old) old_only++; else if(new) new_only++; else neither++} END {printf "total=%d both_ok=%d old_only=%d new_only=%d neither=%d\n", total,both,old_only,new_only,neither}' "$results/results.tsv"
