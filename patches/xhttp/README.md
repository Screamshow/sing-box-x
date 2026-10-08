# XHTTP HTTP/1.1 + HTTP/2 candidate

Original patch series by FiyeroT <me@fiyero.xyz>, from podkop-engine
`v1.14.2-r12` (`46e2370`), GPL-3.0-or-later. Original headers, messages,
authorship and patch bytes are retained; `manifest.json` pins SHA-256.
Xray-core is credited in the original transport comments. Keep all upstream
GPL notices and the complete corresponding source when distributing.

The reviewed series is 0046–0054, 0057–0064, 0069, 0071–0073, plus the
feature declaration idea from 0068. Required 0031 is limited to
`common/readwait/*`; no unrelated gRPC, QUIC or protocol changes are applied.
0046 promotes already pinned cpuid v2.3.0 from indirect to direct without
adopting Podkop's unrelated dependency forks.

`scripts/prepare-xhttp.py` applies the verified series after the unchanged
1.0.2 hardening adapter. It then removes HTTP/3 implementation/tests and
QUIC TLS conversion helpers, and retains an unconditional rejection stub.
The two HTTP/2 regressions in 0064's mixed review2 file are enabled without
with_quic. 0068 advertises only `transport.xhttp`, `transport.xhttp.http1`
and `transport.xhttp.http2`; urltest and decode-link features were not ported.
The effective files and adapter hashes are recorded in prepared source.

X adds a SETTINGS/PING barrier before publishing an HTTP/2 connection: donor
uploads could race the initial MAX_HEADER_LIST_SIZE and cause a GOAWAY instead
of a session-local error. The regression repeats 20 fresh connections. Its
single-socket assertion pins XMUX max_connections=1; the donor default lazily
opens three sockets, which otherwise made that assertion timing-dependent.
The additional helper and test hashes are included in the effective manifest.

No changes to browser TLS fingerprints, REALITY client version, sing-vmess,
or the 1.0.2 uTLS state release. XHTTP HTTP headers follow the donor transport;
this does not replace the configured TLS ClientHello fingerprint.
