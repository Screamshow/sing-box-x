# FakeIP series validation, 2026-10-09

Five original patches from podkop-engine v1.14.2-r12 applied without adaptation
on X main with the Vision direct-reader patch: 0011, 0013, 0080, 0081, 0082.
Upstream 1.14.2, Go 1.26.8; protocol set unchanged, QUIC transports excluded.
Patch checksums and preparation identity are pinned. Preparation repeated
successfully without reapplying patches.

## Automated checks

Unit and race tests passed for dns/transport/fakeip, experimental/cachefile,
route and route/rule. These cover expiry, concurrent use, allocation reserves,
restart/crash behavior, old cache format, persistence failures, reset,
shutdown and missing-record warning behavior. The affected packages are added
to the standard build's unit/race checks. x86_64 and ARM64 builds passed.

## Existing OpenWrt VM

OpenWrt 25 VMware x86_64, Linux 6.12.94. Separate core and generated loopback
HTTP/TLS fixtures; no installed package, UCI, firewall or service changes.
Only the child test process received SIGKILL. Installed Forkop and X remained
running, and /usr/bin/sing-box SHA-256 stayed
0f4870e60ff0aefda3c988a1883396bf9b71ff0ba9b88370fc4804fd5bddb63b.

The final probe exited 0:

```text
dns_100_unique=PASS elapsed_ms=78
casefold_http_tls=PASS
clean_restart=PASS
sigkill_reserve=PASS
flush_missing_record_http_tls=PASS
missing_record_without_name=REFUSED
warning_for_missing_record=PASS
```

The 78 ms value is a local sequential smoke test, not a comparative throughput
benchmark. HTTP response body and TLS certificate were verified. Test-only
configuration initially needed correction to use a non-FakeIP default resolver
and route.default_domain_resolver; those setup failures were not successful runs.

## Limits and behavior

QUIC sniff recovery, ARM64 runtime, full installed Forkop lifecycle, prolonged
load and router RAM impact were not measured. Cache write-failure behavior was
tested by the upstream store tests, not by filling/remounting the VM filesystem.
After a write failure, new names continue from RAM; a crash during allocation
beyond the persisted reserve loses the normal protection against address reuse.
Recovery by sniffing works only if the client provides a usable domain.

This changes FakeIP storage and crash semantics, not subscription hot swapping.
No release, mirror sync, installed-core replacement, commit or push was performed.
