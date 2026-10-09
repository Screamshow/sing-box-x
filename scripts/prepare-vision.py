#!/usr/bin/env python3
"""Apply the pinned Vision direct-reader patch to a private dependency copy."""
import argparse
import hashlib
import json
import shutil
import subprocess
from pathlib import Path


def prepare(source: Path) -> None:
    patches = Path(__file__).resolve().parents[1] / "patches/vision"
    manifest = json.loads((patches / "manifest.json").read_text())
    manifest["adapter_sha256"] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    marker = source / ".forkop-vision.json"
    target = source / ".forkop-deps/sing-vmess"
    if marker.exists():
        if json.loads(marker.read_text()) != manifest or not target.is_dir():
            raise ValueError("Vision preparation changed; use a clean source checkout")
        replacement = json.loads(subprocess.check_output(
            ["go", "mod", "edit", "-json"], cwd=source)).get("Replace", [])
        if not any(r["Old"]["Path"] == manifest["module"] and
                   r["New"]["Path"] == "./.forkop-deps/sing-vmess" for r in replacement):
            raise ValueError("Vision dependency replacement is missing")
        print("Vision direct reader already prepared")
        return
    if target.exists():
        raise ValueError("refusing to overwrite an existing Vision dependency")
    module = json.loads(subprocess.check_output(
        ["go", "list", "-m", "-json", manifest["module"]], cwd=source))
    if module.get("Version") != manifest["module_version"] or module.get("Replace"):
        raise ValueError("sing-vmess dependency changed; review before patching")
    downloaded = json.loads(subprocess.check_output(
        ["go", "mod", "download", "-json", manifest["module"] + "@" +
         manifest["module_version"]], cwd=source))
    if downloaded.get("Sum") != manifest["module_sum"]:
        raise ValueError("sing-vmess module checksum mismatch")
    patch = patches / manifest["patch"]
    if hashlib.sha256(patch.read_bytes()).hexdigest() != manifest["patch_sha256"]:
        raise ValueError("Vision patch checksum mismatch")
    shutil.copytree(downloaded["Dir"], target)
    for path in target.rglob("*"):
        path.chmod(path.stat().st_mode | 0o200)
    subprocess.run(["git", "apply", "--check", str(patch)], cwd=target, check=True)
    subprocess.run(["git", "apply", str(patch)], cwd=target, check=True)
    subprocess.run(["go", "mod", "edit", "-replace=" + manifest["module"] +
                    "=./.forkop-deps/sing-vmess"], cwd=source, check=True)
    marker.write_text(json.dumps(manifest, indent=2) + "\n")
    print("Prepared pinned Vision direct-reader dependency")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=Path)
    prepare(parser.parse_args().source.resolve())
