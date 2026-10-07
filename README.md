# sing-box X

Compact sing-box build for Forkop X on OpenWrt. Release **1.0.0** is based on upstream **1.14.2**.

## Features

Retains uTLS, Clash API, lightweight gRPC transport, WebSocket, HTTP transports and the system TUN stack. Omits gVisor, QUIC-based transports (including Hysteria/Hysteria2/TUIC) and the separate native management API. QUIC sniffing and blocking remain available. Clash API powers Forkop connection monitoring and proxy selection.

The executable is compressed with UPX. Flash storage contains the packed executable; execution decompresses it into RAM. This does not reduce runtime memory in proportion to storage size.

## Packages and installation

Use packaging revision 2 (`1.0.0-r2`). The initial preliminary revision is superseded: its IPK data archive lacked directory entries required by opkg. Revision 2 retains sing-box X version 1.0.0 and corrects package creation.

APK packages target OpenWrt 25; IPK packages target OpenWrt 24. Supported package architectures: `aarch64_cortex-a53` and `x86_64`. Each package owns `/usr/bin/sing-box`, `/etc/init.d/sing-box`, `/etc/config/sing-box` and `/usr/share/sing-box-x/build.json`. The service is disabled by default. Existing configuration is preserved by the package manager.

Packages conflict with `sing-box`, `sing-box-tiny` and `sing-box-extended`. Variant replacement must be managed as a transaction with staged rollback packages. Do not force file overwrites.

Router downloads must use the Forkop mirror:

`https://mirror.51343.ru/forkop/sing-box-x/releases/1.0.0-r2/`

GitHub releases are the mirror's upstream distribution source. Routers do not need GitHub access. `manifest.json` records package hashes, archive bytes, installed payload bytes and packed/plain binary sizes. Installed size describes the packed payload, not its decompressed RAM image; allocator overhead and installation reserve are calculated separately by Forkop.

## Build

On Linux x86-64 install Git, Python 3, Go 1.26.8, curl, tar, xz and an OpenWrt SDK host `apk` with `mkpkg`. Run:

```sh
APK_TOOL=/path/to/openwrt-sdk/staging_dir/host/bin/apk bash scripts/build.sh
```

The recipe pins the upstream commit, Go version, UPX release and UPX checksum. Source preparation fails if expected upstream code changes. The default upstream build still retains its native API; only the X build selects `without_native_api`. `dist/` contains four packages, compressed binary archives, complete prepared source, a manifest and SHA256SUMS.

GitHub Actions runs the same recipe and publishes tagged releases. GitHub APKs are unsigned; the mirror signs its copies using its existing local key and regenerates published hashes. Signing keys never belong in this repository.

## Validation status

ARM64 was installed on a GL-MT6000 running OpenWrt 25.12.5. Forkop startup, local DNS, Clash API and a real HTTPS request through its SOCKS inbound passed; the packed binary allocated 8,344 KiB on writable flash.

Revision 2 APK/IPK installation and reinstall passed in isolated roots on the existing OpenWrt 25/24 x86-64 VMs, including preservation of modified configuration. Published mirror files and APK signatures were verified. See [VALIDATION.md](VALIDATION.md) for exact sizes and scope.

Experimental x86-64 binaries passed clean-boot memory, configuration, DNS, ruleset and basic SOCKS/Clash API tests on existing OpenWrt 24/25 VMs. Real subscription gRPC traffic, latest Xray REALITY compatibility and full package migration checks remain validation items before Forkop canary adoption. Version 1.0.0 (packaging revision 2) is released for the documented feature set. Full Forkop variant migration and rollback checks remain required before changing the Forkop installer default.

## License and source

Based on [SagerNet/sing-box](https://github.com/SagerNet/sing-box), licensed under GPL-3.0-or-later. This repository contains the build recipe and source transformation; each release includes the complete prepared upstream source. The OpenWrt service integration is derived from the OpenWrt sing-box package (GPL-2.0-only).
