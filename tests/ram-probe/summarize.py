#!/usr/bin/env python3
"""Summarize anonymous TSV measurements; fail on missing load or connection errors."""
import csv
import json
import statistics
import sys
from collections import defaultdict
from pathlib import Path

directory = Path(sys.argv[1])
groups = defaultdict(list)
metrics = ("rss_kib", "anon_kib", "shmem_kib", "go_memory_bytes", "fds")
for path in sorted(directory.glob("*.tsv")):
    scenario = path.stem.split("-")[0]
    rounds = defaultdict(list)
    with path.open() as f:
        for row in csv.DictReader(f, delimiter="\t"):
            for key in row:
                if key not in ("label", "phase"):
                    row[key] = int(row[key])
            assert row["errors"] == 0, (path.name, row)
            assert row["active"] == row["target"], (path.name, row)
            assert row["api_connections"] == row["target"], (path.name, row)
            rounds[(row["phase"], row["target"])].append(row)
    for (phase, target), rows in rounds.items():
        version = "1.0.1" if rows[0]["label"].startswith("old") else "1.0.2"
        groups[(scenario, version, phase, target)].append({
            "round": rows[0]["label"], "samples": len(rows),
            **{metric: statistics.mean(r[metric] for r in rows) for metric in metrics},
        })

result = []
for (scenario, version, phase, target), rounds in sorted(groups.items()):
    result.append({"scenario": scenario, "version": version, "phase": phase,
                   "connections": target, "rounds": len(rounds),
                   **{metric: {"median": statistics.median(r[metric] for r in rounds),
                               "min": min(r[metric] for r in rounds),
                               "max": max(r[metric] for r in rounds)} for metric in metrics}})
print(json.dumps(result, indent=2))
