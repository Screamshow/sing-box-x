#!/bin/sh
set -eu
case "${1:-}" in apk|ipk) kind=$1;; *) exit 2;; esac
stage=$(mktemp -d /root/forkop-x-package-probe.XXXXXX)
archive=$(mktemp /tmp/forkop-x-package.XXXXXX)
trap 'rm -f "$archive"; rm -rf "$stage"' EXIT INT TERM
snapshot() {
    if [ "$kind" = apk ]; then apk info -v 2>/dev/null | sort; else opkg list-installed | sort; fi
    sha256sum /etc/config/forkop /etc/config/sing-box /etc/sing-box/config.json
    ubus call service list '{"name":"sing-box"}'
}
snapshot > "$stage/before"
base=https://mirror.51343.ru/forkop/sing-box-x/releases/1.0.0-r2
mkdir -p "$stage/root"
if [ "$kind" = apk ]; then
    curl -fsSL "$base/sing-box-x_1.0.0-r2_x86_64.apk" -o "$archive"
    mkdir -p "$stage/root/lib/apk/db" "$stage/root/etc/apk/keys"
    awk 'BEGIN { RS=""; ORS="\n\n" } $0 !~ /(^|\n)P:sing-box(-tiny|-extended)?(\n|$)/ { print }' /lib/apk/db/installed > "$stage/root/lib/apk/db/installed"
    awk '$0 !~ /^sing-box($|[-=<>~])/' /etc/apk/world > "$stage/root/etc/apk/world"
    curl -fsSL https://mirror.51343.ru/forkop/forkop-apk.pem -o "$stage/root/etc/apk/keys/forkop.pem"
    apk verify --keys-dir "$stage/root/etc/apk/keys" "$archive"
    mv "$archive" "$archive.apk"; archive="$archive.apk"
    if ! apk --root "$stage/root" --no-network --no-scripts --keys-dir "$stage/root/etc/apk/keys" add "$archive" > "$stage/install.log" 2>&1; then cat "$stage/install.log"; exit 1; fi
    apk --root "$stage/root" info -e sing-box-x
else
    curl -fsSL "$base/sing-box-x_1.0.0-2_x86_64.ipk" -o "$archive"
    mv "$archive" "$archive.ipk"; archive="$archive.ipk"
    mkdir -p "$stage/root/usr/lib/opkg" "$stage/root/var/lock" "$stage/root/tmp"
    awk 'BEGIN { RS=""; ORS="\n\n" } $0 ~ /^Package: (ca-bundle|kmod-tun|kernel|libc|libgcc|libgcc1)\n/ { print }' /usr/lib/opkg/status > "$stage/root/usr/lib/opkg/status"
    mkdir -p "$stage/root/usr/lib/opkg/info"
    for name in ca-bundle kmod-tun kernel libc libgcc libgcc1; do
        [ ! -f "/usr/lib/opkg/info/$name.list" ] || cp "/usr/lib/opkg/info/$name.list" "$stage/root/usr/lib/opkg/info/"
    done
    if ! opkg --offline-root "$stage/root" --nodeps install "$archive" > "$stage/install.log" 2>&1; then cat "$stage/install.log"; exit 1; fi
    opkg --offline-root "$stage/root" status sing-box-x
fi
cat "$stage/install.log"
"$stage/root/usr/bin/sing-box" version
printf '%s\n' '{"inbounds":[],"outbounds":[{"type":"direct"}],"experimental":{"clash_api":{"external_controller":"127.0.0.1:19090"}}}' > "$stage/direct.json"
"$stage/root/usr/bin/sing-box" check -c "$stage/direct.json"
wc -c "$stage/root/usr/bin/sing-box"
du -k "$stage/root/usr/bin/sing-box"
test -x "$stage/root/etc/init.d/sing-box"
test -s "$stage/root/etc/config/sing-box"
printf '\n# preserved during reinstall\n' >> "$stage/root/etc/config/sing-box"
sha256sum "$stage/root/etc/config/sing-box" > "$stage/config-before"
if [ "$kind" = apk ]; then
    apk --root "$stage/root" --no-network --no-scripts --keys-dir "$stage/root/etc/apk/keys" add --force-reinstall "$archive" > "$stage/reinstall.log" 2>&1
else
    opkg --offline-root "$stage/root" --nodeps --force-reinstall install "$archive" > "$stage/reinstall.log" 2>&1
fi
cat "$stage/reinstall.log"
sha256sum "$stage/root/etc/config/sing-box" > "$stage/config-after"
cmp "$stage/config-before" "$stage/config-after"
snapshot > "$stage/after"
cmp "$stage/before" "$stage/after"
printf '%s\n' 'PASS: isolated package installation and reinstall, configuration preserved, binary check, global packages/configuration/service state unchanged'
