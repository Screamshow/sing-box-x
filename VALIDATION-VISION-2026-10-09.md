# Vision direct-reader patch: scoped validation

Candidate based on X 1.0.2, upstream 1.14.2, Go 1.26.8. Only FiyeroT/sing-vmess
commit 8bce78ee76de68dee2d879d990c3fbbda3b777f9 is applied to sing-vmess v0.2.8.
No full dependency-fork migration, adaptive Go memory limit or QUIC changes.

Both regression tests pass normally and with the race detector: buffered data
is delivered before the underlying reader is exposed, and the connection
cannot be replaced before direct mode.

## Existing OpenWrt VM

OpenWrt 25 VMware x86_64, Linux 6.12.94, about 986 MiB RAM, no swap.
Installed Forkop and X continued running. Packages, services and configuration
were not replaced. The baseline test binary matched installed X by SHA-256.
Test clients used an isolated loopback mixed inbound and the existing bypass mark.

Local VLESS + TLS + Vision first passed 32/64 HTTPS connections. A subsequent
test used one real TCP VLESS + REALITY + xtls-rprx-vision subscription node and
Firefox uTLS. Subscription credentials and addresses are not included here.
Both variants were UPX packed. Four sequential runs: old1, new1, new2, old2.
Each phase had five seconds settling and ten one-second samples; after closing,
wait twenty seconds plus settling. No forced GC. The inner HTTPS certificate
was verified; requests to generate_204 repeated every five seconds.

All 160 remote-test samples passed active = Clash connection count = target
and load errors = 0. Table entries are medians of two per-run averages.

| Phase | X RSS, MiB | Candidate RSS, MiB | X retained Go, MiB | Candidate retained Go, MiB | X / candidate FD |
| --- | ---: | ---: | ---: | ---: | ---: |
| Idle | 30.81 | 30.81 | 2.39 | 2.40 | 10 / 10 |
| 32 connections | 36.08 | 36.02 | 6.40 | 6.20 | 75 / 139 |
| 64 connections | 39.49 | 37.74 | 9.50 | 7.59 | 139 / 267 |
| Closed | 40.69 | 38.66 | 10.74 | 8.54 | 11 / 11 |

At 64 connections, client RSS decreased by about 1.75 MiB and retained Go
memory by about 1.90 MiB. At 32 connections the RSS change was negligible.
FD increased by about two per connection and returned to baseline after closing.
Kernel pipe memory is excluded from RSS/Go metrics: these measurements do not
establish an equal reduction in total router memory. Retained Go is
StackInuse + HeapInuse + HeapIdle - HeapReleased, not HeapAlloc.

On x86_64 the test binary remained 31,350,946 bytes unpacked. UPX -9 produced
9,938,760 bytes versus baseline 9,938,872 bytes. The 112-byte difference is
not a meaningful size improvement; test version strings and replacement paths
also influence compression. ARM64 and package size were not measured in this probe.

## Limits

One remote node, x86_64, mixed ingress, low-volume HTTPS and two runs each.
No throughput/CPU, ARM64 runtime, TUN/TPROXY ingress, long-term stability or
total system memory claim. REALITY compatibility is confirmed for that node,
not every server version. Test credentials and private logs were removed from
the VM; installed binary hash and running service states remained unchanged.

Anonymous measurements and test-only sources remain in ignored
work/vision-probe. Publication and mirror synchronization are separate steps.
