# Preparation validation â€” 2026-10-07

Release: **1.0.0-r2**, application version **1.0.0**, upstream **1.14.2**.

GitHub build: https://github.com/Screamshow/sing-box-x/actions/runs/37649819963

Mirror manifest: https://mirror.51343.ru/forkop/sing-box-x/releases/1.0.0-r2/manifest.json

| Architecture | Packed executable bytes | APK installed payload bytes | IPK installed payload bytes |
|---|---:|---:|---:|
| aarch64_cortex-a53 | 8534028 | 8536284 | 8536149 |
| x86_64 | 9858380 | 9860624 | 9860489 |

- Both architecture builds and UPX integrity tests passed.
- All seven assets were downloaded through the public mirror and matched published size and SHA-256 records; SHA256SUMS also matched.
- Both mirror APK signatures passed verification with the existing published mirror key.
- Both IPK archives passed directory-layout, executable-mode, ELF-architecture and binary-size checks.
- x86-64 APK installation and reinstall passed in an isolated package root on the existing OpenWrt 25 VM.
- x86-64 IPK installation and reinstall passed in an isolated package root on the existing OpenWrt 24 VM.
- Installed binaries reported 1.14.2-x-1.0.0 and accepted a direct-outbound configuration with Clash API.
- Modified /etc/config/sing-box files survived reinstall in both formats.
- Global package lists, Forkop/sing-box configuration hashes and the running sing-box service state matched before and after.

These package probes use isolated roots and copied package database state. IPK uses --nodeps to avoid downloading dependencies into the isolated root. Actual variant replacement, dependency staging, rollback, real gRPC traffic, latest Xray REALITY and ARM64 execution remain gates before Forkop canary adoption.

The initial preparation revision was superseded after the OpenWrt 24 probe found missing IPK directory entries. Revision 2 fixed the package layout; the application version and executable feature set did not change. Published historical assets were not rewritten.

## ARM64 router verification

Installed the same signed mirror APK on a GL-MT6000 running OpenWrt 25.12.5. The executable accepted the real Forkop configuration and started under procd. Clash API returned the X version, local DNS resolved, and an HTTPS request through the configured SOCKS inbound returned HTTP 204. The generated configuration included a gRPC outbound; no isolated end-to-end gRPC assertion was made.

The packed executable allocated 8,344 KiB (8.15 MiB) on writable flash. Existing Forkop and sing-box UCI hashes matched after startup. The updater's initial premature rollback checks were corrected to wait for asynchronous startup; no binary defect was identified.

Application version remains 1.0.0; packaging revision 2 is released under Git tag 1.0.0-r2. The mirror channel catalog records current stable/preliminary state; immutable release manifests retain their original publication state.
