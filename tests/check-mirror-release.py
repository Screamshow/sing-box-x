#!/usr/bin/env python3
"""Validate the public latest catalog, release manifest and all mirrored assets."""
import argparse
import hashlib
import json
import urllib.request
from pathlib import Path

BASE = "https://mirror.51343.ru/forkop/sing-box-x"


def fetch(url):
    with urllib.request.urlopen(urllib.request.Request(url, headers={"User-Agent": "Forkop-X-validation"}), timeout=120) as response:
        return response.read()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path)
    parser.add_argument("--version", default="1.0.1")
    parser.add_argument("--revision", type=int, default=1)
    args = parser.parse_args()
    catalog = json.loads(fetch(f"{BASE}/latest.json"))
    assert catalog["name"] == "sing-box-x"
    assert catalog["version"] == args.version and catalog["package_revision"] == args.revision
    assert catalog["release_tag"] == args.version and catalog["status"] == "stable"
    assert catalog["upstream_version"] == "1.14.2"
    manifest_url = f"{BASE}/releases/{args.version}/manifest.json"
    assert catalog["manifest_url"] == manifest_url
    manifest = json.loads(fetch(manifest_url))
    for key in ("name", "version", "package_revision", "upstream_version", "upstream_commit", "release_tag", "assets"):
        assert manifest[key] == catalog[key], key
    assert len(manifest["assets"]) == 7
    args.directory.mkdir(parents=True, exist_ok=True)
    (args.directory / "latest.json").write_text(json.dumps(catalog, indent=2) + "\n")
    (args.directory / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    packages = set()
    for asset in manifest["assets"]:
        name = asset["name"]
        assert name == Path(name).name
        url = f"{BASE}/releases/{args.version}/{name}"
        assert asset["url"] == url
        if asset.get("package"):
            assert asset["package"] == "sing-box-x"
            assert asset["architecture"] in ("aarch64_cortex-a53", "x86_64")
            assert asset["format"] in ("apk", "ipk")
            revision = ("r" if asset["format"] == "apk" else "") + str(args.revision)
            assert asset["version"] == f"{args.version}-{revision}"
            assert name == f"sing-box-x_{asset['version']}_{asset['architecture']}.{asset['format']}"
            assert set(asset["dependencies"]) == {"ca-bundle", "kmod-tun"}
            if asset["format"] == "apk":
                assert asset["signed"] is True
            packages.add((asset["architecture"], asset["format"]))
        data = fetch(url)
        assert len(data) == asset["size"], name
        assert hashlib.sha256(data).hexdigest() == asset["sha256"], name
        (args.directory / name).write_bytes(data)
        print(f"PASS: {name}: mirror URL, {len(data)} bytes, SHA-256")
    assert packages == {(arch, kind) for arch in ("aarch64_cortex-a53", "x86_64") for kind in ("apk", "ipk")}
    print("PASS: latest catalog, release manifest, both package architectures/formats and all seven assets")


if __name__ == "__main__":
    main()
