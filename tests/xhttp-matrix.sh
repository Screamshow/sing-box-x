#!/bin/sh
# Forkop must be stopped by the caller: its watchdog owns sing-box processes.
set -eu
umask 077
SB=$1 XRAY=$2 W=$3 TESTS=$4
mkdir -p "$W/www"
printf 'XHTTP-matrix-OK\n' > "$W/www/index.html"
chmod 755 "$W/www"
chmod 644 "$W/www/index.html"
"$SB" generate tls-keypair example.test > "$W/tls.pem"
"$XRAY" x25519 > "$W/reality.keys"
pids=
cleanup() { for pid in $pids; do kill "$pid" 2>/dev/null || true; done; for pid in $pids; do wait "$pid" 2>/dev/null || true; done; pids=; }
trap cleanup EXIT INT TERM
uhttpd -f -p 127.0.0.1:18090 -h "$W/www" > "$W/http.log" 2>&1 & http=$!
trap 'cleanup; kill "$http" 2>/dev/null || true' EXIT INT TERM
echo 'security mode download status' > "$W/results.tsv"
for security in ${XHTTP_TEST_SECURITIES:-plain h1 h2 reality}; do
  for mode in ${XHTTP_TEST_MODES:-auto packet-up stream-up stream-one}; do
    for download in 0 1; do
      [ "$download" = 0 ] || { [ "$security" = h2 ] || [ "$security" = reality ]; } || continue
      [ "$download/$mode" != 1/stream-one ] || continue
      ucode "$TESTS/xhttp-matrix.uc" "$W" "$security" "$mode" "$download"
      "$SB" check -c "$W/client.json" > "$W/check.log" 2>&1
      "$XRAY" run -test -config "$W/server.json" > "$W/server-check.log" 2>&1
      "$SB" run -c "$W/target.json" > "$W/target.log" 2>&1 & pids="$pids $!"
      "$XRAY" run -config "$W/server.json" > "$W/server.log" 2>&1 & pids="$pids $!"
      "$SB" run -c "$W/client.json" > "$W/client.log" 2>&1 & pids="$pids $!"
      sleep 1
      result=passed
      for round in 1 2 3; do
        if ! curl -fsS --max-time 10 --noproxy '' -x http://127.0.0.1:19444 http://127.0.0.1:18090/ > "$W/response" 2> "$W/curl.log" || ! grep -q '^XHTTP-matrix-OK$' "$W/response"; then result=failed; break; fi
      done
      echo "$security $mode $download $result" | tee -a "$W/results.tsv"
      if [ "$result" != passed ]; then cp "$W/client.log" "$W/failed-$security-$mode-$download.log"; fi
      cleanup
    done
  done
done
kill "$http"; wait "$http" 2>/dev/null || true
http=
trap cleanup EXIT INT TERM
