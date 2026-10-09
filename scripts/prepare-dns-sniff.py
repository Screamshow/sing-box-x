#!/usr/bin/env python3
"""Apply the pinned DNS NAT, TLS sniffing and bounded DNS logging series."""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path


def prepare(source: Path) -> None:
    patches = Path(__file__).resolve().parents[1] / "patches/dns-sniff"
    manifest = json.loads((patches / "manifest.json").read_text())
    manifest["adapter_sha256"] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    marker = source / ".forkop-dns-sniff.json"
    if marker.exists():
        if json.loads(marker.read_text()) != manifest:
            raise ValueError("DNS/sniff preparation changed; use a clean source checkout")
        print("DNS/sniff series already prepared")
        return
    for entry in manifest["patches"]:
        patch = patches / entry["file"]
        if hashlib.sha256(patch.read_bytes()).hexdigest() != entry["sha256"]:
            raise ValueError(f"DNS/sniff patch checksum mismatch: {patch.name}")
        subprocess.run(["git", "apply", "--check", str(patch)], cwd=source, check=True)
        subprocess.run(["git", "apply", str(patch)], cwd=source, check=True)
    marker.write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Prepared {len(manifest['patches'])} DNS/sniff patches")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    prepare(parser.parse_args().source.resolve())
