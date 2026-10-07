#!/usr/bin/env python3
"""Assert actual Clash API results from the loopback VM integration probe."""
import json
import sys
from pathlib import Path

expected = {
    "chatgpt.com": "domain_suffix=chatgpt.com",
    "persistent.oaistatic.com": "domain_suffix=oaistatic.com",
    "keyword.example": "domain_keyword=keyword",
    "regex.example": r"domain_regex=^regex\.example$",
    "cidr.example": "ip_cidr=127.0.0.1",
    "default.example": "",
    "failed.example": "",
}
for filename in sys.argv[1:]:
    text = Path(filename).read_text(encoding="utf-8-sig")
    assert "PASS: 7 real SOCKS/HTTP connections; global packages/configuration/services unchanged" in text
    snapshot = json.loads(next(line for line in text.splitlines() if line.startswith('{"connections":')))
    connections = {c["metadata"]["host"]: c for c in snapshot["connections"]}
    assert set(connections) == set(expected), connections.keys()
    for host, payload in expected.items():
        connection = connections[host]
        assert connection["rulePayload"] == payload, (host, connection)
        assert connection["metadata"]["destinationIP"] == "", (host, connection)
        assert connection["chains"] == (["VPN-out"] if payload else ["fallback-out"]), (host, connection)
        if not payload:
            assert connection["rule"] == "final", (host, connection)
    print(f"{filename}: PASS all 7 payload/route assertions, missing destination IP, failed branch and default route")
