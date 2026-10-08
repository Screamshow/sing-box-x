# Runtime hardening provenance

The patch files come from FiyeroT/podkop-engine tag `v1.14.2-r12`,
release commit `46e2370`, under GPL-3.0-or-later. Original authorship and
commit messages are retained. `manifest.json` pins each file's SHA-256.

- 0015: bound QUIC ClientHello reassembly and validate CRYPTO frames.
- 0017: unblock gRPC-lite/HTTP2 writes on a failed request.
- 0025: a slow DoH query does not reset a transport serving later queries.
- 0038: upstream protocol input-validation backport `07512b109`.
- 0039: upstream interruption/close backport `d2cf9747d`.
- 0041: isolate sniffer panics and bound packets retained during sniffing.
- 0043: format observer log entries only while subscribers exist.
- 0084: bound panic stack logging and discard stacks from connection metadata.

0084 is adapted to omit unrelated FakeIP/loop logging dependencies: its
router field context and test section differ from the original. X adds an
independent `TestSnifferStackLoggingIsBounded` regression instead.

`scripts/prepare-hardening.py` additionally adapts the uTLS handshake-memory
changes from 0026/0086. It checks the completed TLS connection state before
skipping a repeated handshake, without importing new browser fingerprints,
REALITY client version changes, XHTTP, or the sing-vmess fork. It applies to
the existing upstream source pinned in build-config.json, before X's native
API exclusion and Clash rule-match instrumentation.

New regression tests cover Firefox data transfer and repeated handshakes,
blocked close versus connection registration, live/frozen DoH connections,
and panic-log retention. Their hashes and the adapter's hash are captured in
the prepared source manifest. Build runs unit and race tests for the affected
packages and checks the production dependency graph.

Source: https://github.com/FiyeroT/podkop-engine/tree/v1.14.2-r12/patches/v1.14

The source archive contains the fully modified Go source and the effective
hardening manifest. Keep this provenance, the original patch headers, and
the project's GPL notices when distributing a build.
