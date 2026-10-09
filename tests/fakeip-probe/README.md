# Scoped FakeIP integration probe

From the prepared sing-box source module, build with the pinned Go toolchain:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /path/fakeip-probe /path/tests/fakeip-probe/main.go
```

Run on the existing OpenWrt test VM using a separate candidate and a new temporary
directory. Capture package/service state first, following the workspace instructions.

```sh
./fakeip-probe /tmp/fakeip-candidate /tmp/new-fakeip-test-directory
```

The probe never reads or modifies the installed sing-box/Forkop configuration.
It starts a separate core and loopback HTTP/TLS/DNS fixtures on ports 48153–48158.
The only SIGKILL is against its own child core. It checks 100 unique DNS answers,
case folding, HTTP/TLS routing, clean persistence, allocation after SIGKILL,
Clash FakeIP flush, HTTP Host/TLS SNI recovery and nameless refusal/warning.
Its TLS certificate is generated locally, trusted explicitly by the client;
no private subscription is needed. Routing uses the existing bypass mark.

The candidate DNS config has a fakeip rule for A/AAAA and a normal default
resolver as required by sing-box 1.14. Run with a fresh directory to avoid
counting an earlier test's cache as new allocations. Exit 0 means all checks passed.

With `DNS_NAT_PROBE=1`, it also sends a slow and fast TXT request from the same
UDP source port after priming that NAT session. The local resolver delays the
slow response 600 ms; the fast response must arrive first within 300 ms.
This checks the DNS NAT handler's asynchronous exchange without external DNS.

With `MEM_LIMIT_PROBE=1`, the probe also verifies that the candidate logs the
automatic Go memory limit starting. The core log is recorded at info level.
