# Runtime optimizations validation — 2026-10-09

Candidate includes the previous Vision, FakeIP and DNS/sniff series plus the
three runtime optimizations in patches/runtime-opt. No version bump, release,
mirror publication or installed package replacement was performed.

## Local checks

- Preparation applied successfully and a second invocation was idempotent.
- Go 1.26.8 ordinary and race tests passed for `.`, `./common/listener`,
  `./route/...`, `./dns`, `./common/sniff`, `./dns/transport/fakeip`.
- Tests cover bounded UDP queue burst/overflow, adaptive memory calculations and
  explicit overrides, failed remote updates, loaded lists and HTTP 304/logging.
- Linux amd64 and arm64 builds passed with X production build tags.

## OpenWrt 25 integration

Existing VMware VM 192.168.1.1; package/service state inspected before testing.
Installed sing-box-x remained 1.0.2-r1. Separate candidate and probe under /tmp,
using loopback fixtures only, in /tmp/runtime-opt-test-20261009.

`DNS_NAT_PROBE=1 MEM_LIMIT_PROBE=1` completed with every assertion passing:

- Slow DNS request did not block the fast request on the same UDP NAT session.
- 100 unique FakeIP queries (95 ms in this run), case folding, HTTP/TLS routing.
- Clean restart persistence, SIGKILL allocation reserve, flush recovery via
  HTTP Host/TLS SNI, nameless refusal with a warning.
- Adaptive limit started on each core launch (634/626/625 MB startup targets
  on this approximately 1 GB VM).

After testing, sing-box and Forkop `status` both reported running. Installed
/usr/bin/sing-box SHA-256 remained
`0f4870e60ff0aefda3c988a1883396bf9b71ff0ba9b88370fc4804fd5bddb63b`.
Forkop's `running` action exits nonzero even though its `status` reports running;
the diagnostic command's final exit code therefore does not indicate probe failure.

These quick checks establish functional compatibility. They do not measure CPU
improvement or prove OOM protection on a 128 MB router. The UDP queue trades
additional possible burst memory for fewer drops.
