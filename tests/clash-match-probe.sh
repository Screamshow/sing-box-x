#!/bin/sh
# Run with an explicit stage containing the candidate binary and test config.
# All listeners and traffic stay on loopback; existing services are untouched.
set -eu
stage=$1
case "$stage" in /tmp/forkop-x-match-*) ;; *) exit 2;; esac
core_pid=; http_pid=; request_pids=
cleanup() {
    for pid in $request_pids $core_pid $http_pid; do kill "$pid" 2>/dev/null || true; done
    for pid in $request_pids $core_pid $http_pid; do wait "$pid" 2>/dev/null || true; done
}
trap cleanup EXIT INT TERM
snapshot() {
    if command -v apk >/dev/null; then apk info -v 2>/dev/null | sort; else opkg list-installed | sort; fi
    sha256sum /etc/config/forkop /etc/config/sing-box /etc/sing-box/config.json
    ubus call service list
}
snapshot > "$stage/before"
"$stage/sing-box" version
"$stage/sing-box" check -c "$stage/config.json"
mkdir -p "$stage/www/cgi-bin"
printf '%s\n' '#!/bin/sh' 'sleep 8' 'printf "Content-Type: text/plain\r\n\r\nPASS\n"' > "$stage/www/cgi-bin/hold"
chmod 755 "$stage/www/cgi-bin/hold"
/usr/sbin/uhttpd -f -p 127.0.0.1:18080 -h "$stage/www" > "$stage/http.log" 2>&1 & http_pid=$!
"$stage/sing-box" run -c "$stage/config.json" > "$stage/core.log" 2>&1 & core_pid=$!
ready=0
for i in 1 2 3 4 5 6 7 8 9 10; do
    if curl -fsS http://127.0.0.1:19090/version > "$stage/version.json"; then ready=1; break; fi
    sleep 1
done
test "$ready" = 1 || { cat "$stage/core.log"; exit 1; }
for host in chatgpt.com persistent.oaistatic.com keyword.example regex.example cidr.example default.example failed.example; do
    curl --noproxy '' --socks5-hostname 127.0.0.1:19080 -fsS --max-time 25 "http://$host:18080/cgi-bin/hold" > "$stage/$host.response" 2> "$stage/$host.error" &
    request_pids="$request_pids $!"
done
sleep 2
curl -fsS http://127.0.0.1:19090/connections > "$stage/connections.json"
for pid in $request_pids; do wait "$pid" || { cat "$stage"/*.error "$stage/core.log"; exit 1; }; done
request_pids=
for host in chatgpt.com persistent.oaistatic.com keyword.example regex.example cidr.example default.example failed.example; do
    grep -qx PASS "$stage/$host.response"
done
cleanup
core_pid=; http_pid=
snapshot > "$stage/after"
cmp "$stage/before" "$stage/after"
echo 'PASS: 7 real SOCKS/HTTP connections; global packages/configuration/services unchanged'
cat "$stage/connections.json"
