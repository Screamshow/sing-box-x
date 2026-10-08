#!/usr/bin/env python3
"""Apply the reviewed runtime fixes; keep fingerprints and protocol set unchanged."""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path


def prepare(source: Path) -> None:
    patches = Path(__file__).resolve().parents[1] / "patches"
    manifest = json.loads((patches / "manifest.json").read_text())
    tests = Path(__file__).with_name("hardening-tests")
    manifest["adapter_sha256"] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    manifest["tests"] = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(tests.glob("*.go"))}
    marker = source / ".forkop-hardening.json"
    if marker.exists():
        if json.loads(marker.read_text()) != manifest:
            raise ValueError("hardening manifest changed; use a clean source checkout")
        print("Runtime hardening already prepared")
        return
    for entry in manifest["patches"]:
        patch = patches / entry["file"]
        if hashlib.sha256(patch.read_bytes()).hexdigest() != entry["sha256"]:
            raise ValueError(f"patch digest mismatch: {patch.name}")
        subprocess.run(["git", "apply", "--check", str(patch)], cwd=source, check=True)
        subprocess.run(["git", "apply", str(patch)], cwd=source, check=True)
    prepare_tls_memory(source)
    for test in tests.glob("*.go"):
        destination = source / test.name.replace("__", "/")
        if destination.exists():
            raise ValueError(f"refusing to overwrite test: {destination}")
        destination.write_bytes(test.read_bytes())
    marker.write_text(json.dumps(manifest, indent=2) + "\n")
    print(f"Applied {len(manifest['patches'])} reviewed runtime patches")


def prepare_tls_memory(source: Path) -> None:
    """Adapt podkop 0026/0086 without changing the REALITY or browser Hello."""
    path = source / "common/tls/utls_client.go"
    text = path.read_text()
    old = "func (c *utlsALPNWrapper) HandshakeContext(ctx context.Context) error {\n"
    if text.count(old) != 1:
        raise ValueError("uTLS wrapper changed")
    text = text.replace(old, old + "\tif c.Conn.ConnectionState().HandshakeComplete {\n\t\treturn c.UConn.HandshakeContext(ctx)\n\t}\n")
    old = "\treturn c.UConn.HandshakeContext(ctx)\n}\n\nfunc NewUTLSClient"
    if text.count(old) != 1:
        raise ValueError("uTLS handshake changed")
    text = text.replace(old, """\terr := c.UConn.HandshakeContext(ctx)
\tif err != nil {
\t\treturn err
\t}
\treleaseHandshakeState(c.UConn)
\treturn nil
}

// Adapted from FiyeroT/podkop-engine patches 0026/0086 (GPL-3.0-or-later).
// Drop objects used only during the completed handshake, retaining the TLS Conn.
func releaseHandshakeState(uConn *utls.UConn) {
\tuConn.HandshakeState.Hello = nil
\tuConn.HandshakeState.ServerHello = nil
\tuConn.HandshakeState.State13 = utls.TLS13OnlyState{}
\tuConn.Extensions = nil
}

func NewUTLSClient""")
    path.write_text(text, newline="\n")
    path = source / "common/tls/reality_client.go"
    text = path.read_text()
    old = "\treturn &realityClientConnWrapper{uConn}, nil"
    if text.count(old) != 1:
        raise ValueError("REALITY wrapper changed")
    path.write_text(text.replace(old, "\treleaseHandshakeState(uConn)\n" + old), newline="\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    prepare(parser.parse_args().source.resolve())
