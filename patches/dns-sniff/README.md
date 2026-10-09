# DNS NAT, TLS sniffing and bounded DNS logs

Four original GPL-3.0-or-later patches from FiyeroT/podkop-engine v1.14.2-r12,
preserved without adaptation, with authorship and pinned SHA-256:

- 0008: handle DNS packets from inbound UDP NAT sessions without persistent
  per-session DNS reader goroutines and packet queues.
- 0012: exchange off the inbound reading loop, so a slow query does not block
  unrelated packets on a reused session.
- 0018: stop the TLS server handshake after capturing ClientHello; sniffing
  does not proceed into an unnecessary key exchange.
- 0027: log the first DNS failure per server, then a count/last error every
  ten seconds, excluding caller cancellation and duplicate route logging.

The dependency order is retained. No provider-DNS fallback, UDP socket policy,
DoT multiplexer, TLS read-waiter, adaptive memory limit or URLTest changes.
The source archive includes original patches, preparation manifest and adapter.
See VALIDATION-DNS-SNIFF-2026-10-09.md for the scoped checks.
