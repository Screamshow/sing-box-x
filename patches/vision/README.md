# Vision direct-reader optimization

Original patch by FiyeroT, commit
`8bce78ee76de68dee2d879d990c3fbbda3b777f9` in FiyeroT/sing-vmess.
The original author, commit message and two regression tests are preserved.

Only this patch is applied to upstream sing-vmess v0.2.8. After Vision enters
direct read mode and drains its buffered data, the copy loop may use the
underlying reader, including read waiters or splice. Writes remain through
Vision. The connection remains non-replaceable before direct mode.

Preparation verifies the module and patch checksums, copies the dependency
into the prepared source tree and adds a relative Go module replacement.
The source archive therefore includes the effective dependency, its LICENSE,
the patch and provenance. The shared Go module cache is never modified.

Build runs the dependency's tests, including the two added tests, normally
and with the race detector. See VALIDATION-VISION-2026-10-09.md for the scoped
OpenWrt integration measurements.
