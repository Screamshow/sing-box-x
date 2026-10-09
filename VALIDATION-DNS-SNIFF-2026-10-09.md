# DNS and TLS sniffing validation, 2026-10-09

Scope: original podkop-engine v1.14.2-r12 patches 0008, 0012, 0018 and 0027,
applied without adaptation on the local X candidate containing Vision and FakeIP.
Upstream 1.14.2, Go 1.26.8, existing X tags and protocol scope unchanged.
Preparation/checksums and a repeated idempotent preparation passed.

Unit and race tests passed for dns, route, route/rule, common/sniff and
dns/transport/fakeip. Added upstream tests cover browser/Go ClientHello inputs,
fragmented/truncated hello, non-TLS/no-SNI input, first DNS failure, suppression,
per-server state, interval summaries and cancellation. These packages are
included in the standard build test list. x86_64 and ARM64 builds passed.

## Existing OpenWrt 25 VM

Separate candidate on VMware x86_64; installed Forkop and X continued running.
No package, UCI, firewall or installed service changes. Before the probe,
package/service state was checked. A temporary local DNS resolver on port 48158
returns a slow TXT answer after 600 ms and a fast answer immediately. Three
queries reuse one UDP source port: prime, slow, fast. The fast answer arrived
first and within 300 ms, demonstrating that the slow existing-session exchange
does not block the inbound read loop.

```text
same_udp_session_slow_does_not_block_fast=PASS
dns_100_unique=PASS elapsed_ms=82
casefold_http_tls=PASS
clean_restart=PASS
sigkill_reserve=PASS
flush_missing_record_http_tls=PASS
missing_record_without_name=REFUSED
warning_for_missing_record=PASS
```

The probe exited 0. HTTP body and generated TLS certificate were verified,
including SNI recovery after FakeIP flush. SIGKILL targeted only the child
candidate. The elapsed DNS value is a smoke test, not a comparative benchmark.
Bounded DNS logging was tested by unit/race tests, not a prolonged live outage.
No claimed CPU/RAM improvement was measured here, and ARM64 runtime is untested.

Afterwards both installed services remained running and /usr/bin/sing-box hash
was unchanged: 0f4870e60ff0aefda3c988a1883396bf9b71ff0ba9b88370fc4804fd5bddb63b.
No commit, push, release or mirror update was performed.
