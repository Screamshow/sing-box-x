# TLS/REALITY fingerprint provenance

Original patches 0001, 0002, 0010 and 0085 by FiyeroT, from podkop-engine
v1.14.2-r12 (46e2370), GPL-3.0-or-later. Headers and bytes are retained;
manifest.json pins SHA-256. The donor credits Xray-core and
refraction-networking/utls aa6edf4 for Firefox 148 / Safari 26.3 presets.

prepare-tls-fingerprints.py applies the coherent series after unchanged
hardening/uTLS memory release and before XHTTP. No dependency forks or
HTTP/3/QUIC transport code are imported. Prepared source records adapter,
original patch hashes and effective TLS file hashes. This deliberately
changes Firefox/Safari/random/randomized and REALITY client version/share;
older reports describing unchanged fingerprints are historical baselines.
