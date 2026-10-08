#!/usr/bin/env python3
import hashlib,json,subprocess,os,io,tarfile
from pathlib import Path
root=Path(__file__).resolve().parents[1];c=json.loads((root/'build-config.json').read_text());out=root/'dist';assets=[]
for f in sorted(out.iterdir()):
    if not f.is_file() or f.name in ('manifest.json','SHA256SUMS'):continue
    asset={'name':f.name,'sha256':hashlib.sha256(f.read_bytes()).hexdigest(),'size':f.stat().st_size,'url':f"{c['mirror_base']}/releases/{c.get('release_tag',c['version'])}/{f.name}"}
    for goarch,arch in [('arm64','aarch64_cortex-a53'),('amd64','x86_64')]:
        if arch in f.name and f.suffix in ('.apk','.ipk'):
            installed=0
            if f.suffix=='.ipk':
                with tarfile.open(f, 'r:gz') as outer:
                    data=outer.extractfile('data.tar.gz').read()
                with tarfile.open(fileobj=io.BytesIO(data), mode='r:gz') as payload:
                    installed=sum(member.size for member in payload if member.isfile())
            if f.suffix=='.apk':
                dump=subprocess.check_output([os.environ['APK_TOOL'],'adbdump',str(f)],text=True)
                import re
                match=re.search(r'installed-size:\s*(\d+)',dump)
                if not match:raise ValueError('APK installed size unavailable')
                installed=int(match.group(1))
            asset.update(package='sing-box-x',version=f"{c['version']}-{'r' if f.suffix=='.apk' else ''}{c['package_revision']}",architecture=arch,format=f.suffix[1:],installed_size=installed,binary_bytes=(root/'work'/f'sing-box-{goarch}').stat().st_size,plain_binary_bytes=(root/'work'/f'sing-box-{goarch}.plain').stat().st_size,dependencies=['ca-bundle','kmod-tun'])
    assets.append(asset)
manifest={'schema':1,'name':'sing-box-x','version':c['version'],'package_revision':c['package_revision'],'upstream_version':c['upstream_version'],'upstream_commit':c['upstream_commit'],'release_tag':c.get('release_tag',c['version']),'status':c.get('release_status','candidate'),'assets':assets}
(out/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(f.read_bytes()).hexdigest()+'  '+f.name+'\n' for f in sorted(out.iterdir()) if f.is_file() and f.name!='SHA256SUMS'))
