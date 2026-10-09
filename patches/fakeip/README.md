# In-memory FakeIP series

Original FiyeroT/podkop-engine patches from v1.14.2-r12, preserved without
modification, with author headers and SHA-256 pins in manifest.json:

- 0011: memory table, background persistence, expiry and crash allocation reserve.
- 0013: recover a missing FakeIP record from sniffed HTTP/TLS/QUIC domain.
- 0080: continue allocating in memory while cache writes fail.
- 0081: stop the background loop before refusing saves during shutdown.
- 0082: log a missing-record refusal as an explanatory warning.

The existing X input-validation backport (0038) remains applied. No unrelated
DNS fallback, URLTest or memory-limit patches are included. The source archive
includes these originals, the prepared sources, manifest and adapter.

The cache-write failure fallback is intentionally weaker: a crash while
allocating beyond the persisted reserve is not protected against address reuse.
Addresses whose mappings reached the cache retain the normal guarantees.
Sniff-based recovery requires a sniff rule and an identifiable client domain;
it does not recover nameless protocols. QUIC recovery has not been tested here.

See VALIDATION-FAKEIP-2026-10-09.md and tests/fakeip-probe for scoped validation.
