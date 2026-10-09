# Quick VM measurements and private subscription validation

Existing VMware x86_64 VMs: OpenWrt 25.12.5 (986 MiB RAM) and 24.10.8
(202 MiB RAM), no swap. Candidate includes Vision, FakeIP, DNS/sniff and runtime
optimization series. Baseline is installed X 1.0.2, matching the known binary
SHA-256. Candidate was UPX -9 packed (9,972,816 bytes; unpacked 31,445,154).
No installed packages, service configuration or firewall rules were changed.

## Subscription

Downloaded the user-authorized subscription directly on both VMs using the
normal INCY headers and Forkop's installed parser. It contained 58 TCP VLESS
REALITY Vision nodes, including two Safari and 56 Firefox uTLS profiles, plus
other transports outside this test's scope. No addresses, UUIDs, public keys,
subscription URLs or private logs are included in this report or evidence.

- OpenWrt 25: paired baseline/candidate HTTPS requests returned 204 for all
  10 sampled Firefox nodes and both Safari nodes (12/12 for each build).
- OpenWrt 24: sequential candidate HTTPS requests returned 204 for the same
  12 nodes, avoiding simultaneous additional cores on the small VM.
- Separate candidate FakeIP/DNS integration passed on both VMs: 100 unique
  queries, case folding, HTTP/TLS routing, clean restart, SIGKILL reserve,
  flush/missing-record recovery, nameless refusal and warning, same-session
  asynchronous DNS exchange, automatic memory limit startup.
- Subscription data, generated private configs and private logs were removed
  from both VMs after testing. Temporary candidate/helper binaries were removed.

## CPU and memory

Same first REALITY Vision node for old/new measurements, Firefox uTLS, mixed
loopback inbound, verified inner HTTPS certificate, persistent requests to
generate_204 every five seconds. Separate fresh processes, baseline then
candidate on each VM; no forced GC. Two-second settling and five one-second
samples per phase, five-second delay after closing. This is a short low-volume
test, not a throughput benchmark. Measurement helpers and compressed candidate
resided in tmpfs; installed Forkop and sing-box remained active throughout.

| VM | Connections | Baseline RSS MiB | Candidate RSS MiB | Baseline/candidate retained Go MiB | Baseline/candidate CPU % of one core |
| --- | ---: | ---: | ---: | --- | --- |
| 25 | 32 | 36.00 | 35.88 | 6.65 / 6.05 | 0.75 / 1.25 |
| 25 | 64 | 38.70 | 37.45 | 9.08 / 7.16 | 1.00 / 1.24 |
| 24 | 32 | 35.79 | 35.95 | 6.10 / 6.17 | 0.74 / 0.99 |
| 24 | 64 | 38.93 | 37.39 | 9.18 / 7.17 | 1.00 / 1.24 |

CPU uses deltas of /proc/PID/stat user+system ticks with Linux USER_HZ=100
and elapsed sample timestamps. Low counts and short sampling do not establish
a CPU improvement or a meaningful regression. Go memory is Clash's retained
runtime metric, not HeapAlloc or total VM memory. RSS excludes kernel pipe/socket
memory. Candidate FD count is 139/267 at 32/64 connections versus baseline 75/138,
returning to 11 after closing; idle FD count is 10.

### OpenWrt 24 initial OOM

The first candidate run was killed by the kernel during the 64-connection phase
(PID 12232, anon RSS 24,952 KiB, shmem RSS 12,036 KiB); baseline completed that
run. Installed services continued running. This failed run is excluded from the
successful table, not treated as a pass. A candidate run limited to 32 passed;
then a complete 32/64 run passed after removing an unused test helper from
tmpfs. The table uses that complete retry. Different memory headroom means this
is not a strict identical-environment comparison on VM 24. The automatic Go
limit is a soft target and did not prevent the initial system OOM.

At 64 connections the successful runs show about 1.25–1.54 MiB lower process RSS
and about 2 MiB lower retained Go memory. Total-system memory benefit and
stability under exhausted RAM remain unproven. No ARM64 runtime claim.

Anonymous phase summaries: tests/evidence/vm-quick-2026-10-09.json. These include
the separate 32-only successful run and complete retry, with explicit labels.
Test-only helper sources/raw anonymous samples remain in ignored work/quick-vm-probe.

Final status on both VMs: Forkop and sing-box running; installed binary SHA-256
unchanged: 0f4870e60ff0aefda3c988a1883396bf9b71ff0ba9b88370fc4804fd5bddb63b.
No commit, push, version bump, release or mirror synchronization was performed.
