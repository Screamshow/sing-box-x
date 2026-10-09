# Runtime optimizations

Pinned GPL-3.0-or-later patches from FiyeroT/podkop-engine v1.14.2-r12.
The manifest records the original patch hashes. Preparation verifies them and
records the adapter hash; rerunning against the same source is idempotent.

- 0014: raise the bounded UDP output queue from 64 to 1024 packets. This absorbs
  short bursts; it can retain more packet buffers during a burst and is not a RAM
  reduction. Full queues still drop immediately rather than blocking writers.
- 0032: Linux Go soft memory limit follows available system/cgroup memory after
  GC, reserving resident code pages and a minimum useful heap margin. Explicit
  GOMEMLIMIT or experimental.debug.memory_limit disables the automatic policy.
  This is a soft GC target, not a process RSS cap or an OOM guarantee.
- 0045: collect garbage after a genuinely loaded remote rule-set, rather than
  failed updates or HTTP 304. Repeated failures get bounded error logging.

0045 is adapted by scripts/prepare-runtime-opt.py because its original context
depends on podkop's start-with-empty-lists patch. X keeps its existing startup
and scheduling behavior; only loaded-list tracking and failure logging are
ported. The upstream test fixture is replaced with an equivalent local struct
because its helper belongs to that excluded patch.

Original patches and the adapter are included in corresponding-source archives
by scripts/archive-provenance.py.
