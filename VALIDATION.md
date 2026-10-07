# X 1.0.1 validation - 2026-10-07


Stable release: https://github.com/Screamshow/sing-box-x/releases/tag/1.0.1

Successful build and race checks:
https://github.com/Screamshow/sing-box-x/actions/runs/37664492872

Mirror catalog: https://mirror.51343.ru/forkop/sing-box-x/latest.json

Release manifest:
https://mirror.51343.ru/forkop/sing-box-x/releases/1.0.1/manifest.json

- X application 1.0.1, upstream 1.14.2, APK 1.0.1-r1, IPK 1.0.1-1.
- All CI asset hashes and SHA256SUMS passed verification before publication.
- The final CI x86-64 binary is byte-for-byte identical to the candidate run
  on both VMs: SHA-256 `ef461190a690c2643982b95839d429e75e7acdd72c8e48deff72b814cd91a410`.
- The existing mirror publisher verified GitHub asset digests, signed both APKs
  using the existing mirror identity and verified both signatures.
- Public latest.json and manifest agree on release/package identity; all seven
  public assets passed mirror URL, length and SHA-256 validation.
- Signed x86-64 APK installation/reinstallation passed in an isolated package
  root on OpenWrt 25. x86-64 IPK installation/reinstallation passed in an
  isolated root on OpenWrt 24. Both preserved the modified UCI conffile and
  accepted the test configuration. Global packages, configs and running
  service snapshots were unchanged.
- Temporary probe processes and directories were removed from both VMs.
- ARM64 was built, UPX-tested, packaged and signature/hash-verified; its new
  binary was not installed or executed on the live ARM64 router in this task.

| Architecture | UPX binary bytes | APK installed payload bytes |
|---|---:|---:|
| aarch64_cortex-a53 | 8,537,372 | 8,539,655 |
| x86_64 | 9,863,528 | 9,865,799 |

The release uses the successful manual CI build's immutable assets. Creating
the release also triggered a redundant tag build; that duplicate was cancelled
because its publish step would attempt to recreate the existing release.
The monitoring formatter and rebuilt LuCI bundle are included in Forkop main;
they were not deployed to 192.168.90.1 as part of this release task.

## Historical X 1.0.0 validation

# Preparation validation â€” 2026-10-07

Release: **1.0.0**, package revision **2**, upstream **1.14.2**. The stable release reuses the validated revision 2 executables and packages without rebuilding them.

GitHub build: https://github.com/Screamshow/sing-box-x/actions/runs/37649819963

Mirror manifest: https://mirror.51343.ru/forkop/sing-box-x/releases/1.0.0/manifest.json

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
