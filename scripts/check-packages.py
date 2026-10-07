#!/usr/bin/env python3
"""Validate IPK directory layout before release publication."""
import io,json,tarfile
from pathlib import Path
root=Path(__file__).resolve().parents[1]
for package in sorted((root/'dist').glob('*.ipk')):
    with tarfile.open(package,'r:gz') as outer:
        assert set(outer.getnames())=={'debian-binary','control.tar.gz','data.tar.gz'},package.name
        assert outer.extractfile('debian-binary').read()==b'2.0\n'
        data=outer.extractfile('data.tar.gz').read()
    with tarfile.open(fileobj=io.BytesIO(data),mode='r:gz') as archive:
        members={member.name.removeprefix('./').rstrip('/'):member for member in archive}
        for name in ('etc','etc/config','etc/init.d','usr','usr/bin','usr/share','usr/share/sing-box-x'):
            assert name in members and members[name].isdir(),(package.name,name)
        binary=members['usr/bin/sing-box'];assert binary.isfile() and binary.mode==0o755
        header=archive.extractfile(binary).read(20);assert header[:5]==b'\x7fELF\x02'
        machine=int.from_bytes(header[18:20],'little')
        assert machine==(183 if 'aarch64' in package.name else 62)
        assert members['etc/init.d/sing-box'].mode==0o755
        metadata=json.load(archive.extractfile(members['usr/share/sing-box-x/build.json']))
        assert metadata['binary_bytes']==binary.size
    print(package.name,'directory layout, executable mode, ELF architecture and binary size OK')
