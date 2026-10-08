#!/usr/bin/env python3
"""Apply the coherent podkop TLS/REALITY fingerprint series after hardening."""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path


def prepare(source: Path) -> None:
    patches = Path(__file__).resolve().parents[1] / 'patches/tls-fingerprints'
    manifest = json.loads((patches / 'manifest.json').read_text(encoding='utf-8-sig'))
    manifest['adapter_sha256'] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    marker = source / '.forkop-tls-fingerprints.json'
    if marker.exists():
        if json.loads(marker.read_text()) != manifest:
            raise ValueError('TLS fingerprint manifest changed; prepare fresh source')
        return
    for entry in manifest['patches']:
        patch = patches / entry['file']
        if hashlib.sha256(patch.read_bytes()).hexdigest() != entry['sha256']:
            raise ValueError(f'TLS patch digest mismatch: {patch.name}')
        subprocess.run(['git', 'apply', '--check', str(patch)], cwd=source, check=True)
        subprocess.run(['git', 'apply', str(patch)], cwd=source, check=True)
    subprocess.run(['gofmt', '-w', 'common/tls'], cwd=source, check=True)
    marker.write_text(json.dumps(manifest, indent=2) + '\n')
    manifest['effective_files'] = {
        str(p.relative_to(source)): hashlib.sha256(p.read_bytes()).hexdigest()
        for p in sorted((source / 'common/tls').glob('*')) if p.is_file()
    }
    (source / '.forkop-tls-fingerprints-effective.json').write_text(json.dumps(manifest, indent=2) + '\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source', type=Path)
    prepare(parser.parse_args().source.resolve())
