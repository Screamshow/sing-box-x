# Customer-authorized router quick check

GL-MT6000 at 192.168.90.1, OpenWrt 25.12.5, ARM64, approximately 986 MiB RAM.
Installed X 1.0.2-r1; Forkop and sing-box running. Package/service/memory state
was read before testing. Candidate ran separately from /tmp on loopback ports;
no installed package, service settings, UCI or firewall configuration changed.

Private test profiles were generated on-router from the existing configuration.
It contains 58 TCP VLESS REALITY Vision nodes and an HTTPS DNS server. The
isolated real-node test preserved configured non-FakeIP DNS servers, selected
the HTTPS server as final/domain resolver and disabled DNS cache. It used the
existing bypass routing mark and a separate mixed inbound.

## Results

- Candidate and installed baseline: ordinary certificate-verified HTTPS
  generate_204 requests succeeded on three sampled REALITY Vision nodes through
  the configured DoH resolver (3/3 each).
- Candidate FakeIP/DNS isolated fixtures all passed: unique queries, casefold,
  HTTP/TLS routing, clean restart, SIGKILL reserve, flush recovery, nameless
  refusal/warning, same-UDP-session nonblocking DNS exchange and adaptive memory
  limit startup. 100 unique DNS queries took 143 ms in this run.
- Concurrent persistent HTTPS probe on the same first real node was not clean
  for either build. Candidate reached 27/32 then 55/64 active connections, with
  9 cumulative load errors. Baseline reached 29/32 then 53/64, with 11 errors.
  Core stayed alive. The probe counts SOCKS/TLS/HTTP failures without exposing
  private logs; exact failure stage was not established. This does not prove a
  regression, nor does it establish a clean stress-test pass. No throughput or
  statistically meaningful CPU comparison is claimed from failed load phases.
- Actual DoH end-to-end resolution was exercised; the controlled slow/fast DNS
  assertion still uses a local UDP upstream, not a controlled slow DoH server.

Temporary subscription/node configs and private logs were removed afterward,
along with the uploaded candidate/helper executables. Anonymous TSV samples
remain in ignored work/router-quick-results. Final installed hash unchanged:
1a59b481a1464e61c952fdbfcd354f8575500e7bca6c5ffbc590a3c67e44a14d.
Both production services still reported running. No release was created because
the requested all-clear condition was not fully established by the load test.
