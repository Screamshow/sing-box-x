# Clash API match evidence

X 1.0.1 populates `rulePayload` with a positive inline destination condition
captured during actual route evaluation, for example
`domain_suffix=chatgpt.com` or `ip_cidr=127.0.0.1`.
The existing `rule`, connection metadata, chains and API endpoints are retained.

Supported evidence: domain, domain suffix, keyword, RE2 regex and destination
CIDR. AND combines successful child evidence; OR selects the successful definite
branch. Failed, inverted, deferred and opaque rule-set evaluations publish no
positive evidence. The complete logical rule remains available in `rule`.
Rule-set names continue to be handled by Forkop's existing display logic.
Pure exclusions, device-only rules and binary rule sets do not expose a specific
positive destination entry. Payload strings belong to the connection snapshot,
not shared rule objects or a later browser-side reconstruction.

The source patch adds regression tests to upstream rule and Clash API packages.
The build runs their existing tests and race detector checks before packaging.

For VM integration, copy the candidate executable and `clash-match.json` into
a newly created `/tmp/forkop-x-match-*` directory as `sing-box` and `config.json`,
then run `clash-match-probe.sh <directory>` there. Capture the output and run
`python3 tests/check-clash-match.py <log>` on the operator host. The probe starts
temporary sing-box and HTTP listeners only on loopback, runs seven real SOCKS
requests, cleans up its processes, and compares global package/configuration/
service snapshots. It neither installs packages nor stops current services.
