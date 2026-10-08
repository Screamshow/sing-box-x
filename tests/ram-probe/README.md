# RAM integration probe

Build the standalone helper with the same Go toolchain as X:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o ram-probe ./tests/ram-probe/main.go
```

Run only on the existing test VM. Capture package/service/config state first.
Stop Forkop for the measurement: its watchdog and explicit Stop detect/terminate
additional sing-box processes. Restore the original service afterwards.
Keep private subscription input, generated config, logs and TLS keys in a
0700 temporary directory; none belong in Git.

`prepare DIRECTORY` generates a temporary self-signed local TLS fixture and
VLESS server/client configs. Start `echo` and the fixed baseline X with
`server.json` separately. The measured client uses Firefox uTLS through the
local server, with 32, 128 and 256 connections. Each connection exchanges
64 bytes every second. The fixture has no REALITY or Vision; this isolates
ordinary uTLS connections.

`real-config NORMALIZED_SUBSCRIPTION OUTPUT` selects a TCP VLESS REALITY
Firefox node from subscription JSON normalized by Forkop. Its users and
credentials are never printed. `run` in `real` mode uses 32/64 connections;
each performs HTTPS with certificate verification and repeats a small 204
request every five seconds. This measures the actual remote node, but does
not cover other nodes, transports or long-term performance.

```sh
./ram-probe run /path/to/core client.json measurements.tsv old1 local
./ram-probe run /path/to/core real.json measurements.tsv old1 real gc
```

The helper starts and stops the measured core itself. Configs use loopback
ports 48100–48104. It records process RSS, anonymous/shared pages, file
descriptors, actual Clash connection count and the Clash Go memory metric:
`StackInuse + HeapInuse + HeapIdle - HeapReleased`. This is retained Go
memory, not HeapAlloc or total VM RAM. The local server and load-generator
memory are excluded from the measured PID. PSS is unavailable on the test
kernel; RSS includes shared executable pages.

Phases have five seconds of settling followed by ten one-second samples.
After closing connections, wait twenty seconds plus settling before sampling.
Optional `gc` calls the existing loopback debug GC endpoint and takes three
additional samples after two seconds. Rows with `_gc` are a diagnostic lower
retention measurement, not a simulation of normal production GC. In this
mode later phases follow the preceding forced GC and must be labelled as
such when reported.

Use three fresh processes per version, alternating order:
`old1 new1 new2 old2 old3 new3`. Do not execute both clients at once.
Use anonymous files named `local-old1.tsv`, `real-new1.tsv`, etc.
`summarize.py DIRECTORY` validates that every sampled connection is alive,
Clash reports the requested count and no load errors occurred. It calculates
the mean within each round and median/min/max across rounds. Failed or
watchdog-interrupted runs must be kept outside the summarized directory.
