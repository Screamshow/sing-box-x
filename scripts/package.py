#!/usr/bin/env python3
"""Package the packed executable for OpenWrt without host-architecture guessing."""
import argparse, gzip, hashlib, io, json, os, shutil, stat, subprocess, tarfile, tempfile
from pathlib import Path, PurePosixPath
ROOT=Path(__file__).resolve().parents[1]
def tar_bytes(files):
    stream=io.BytesIO()
    with gzip.GzipFile(fileobj=stream,mode='wb',mtime=0) as gz:
        with tarfile.open(fileobj=gz,mode='w',format=tarfile.GNU_FORMAT) as tar:
            directories={str(parent) for name,_,_ in files for parent in PurePosixPath(name).parents if str(parent)!='.'}
            for name in sorted(directories,key=lambda value:(value.count('/'),value)):
                item=tarfile.TarInfo(name+'/');item.type=tarfile.DIRTYPE;item.mode=0o755
                item.uid=item.gid=0;item.mtime=int(os.environ.get('SOURCE_DATE_EPOCH','0'));tar.addfile(item)
            for name,content,mode in sorted(files):
                item=tarfile.TarInfo(name);item.size=len(content);item.mode=mode
                item.uid=item.gid=0;item.mtime=int(os.environ.get('SOURCE_DATE_EPOCH','0'))
                tar.addfile(item,io.BytesIO(content))
    return stream.getvalue()
def main():
    p=argparse.ArgumentParser();p.add_argument('--apk-tool',required=True);a=p.parse_args()
    c=json.loads((ROOT/'build-config.json').read_text());v=c['version'];rev=c['package_revision']
    out=ROOT/'dist';out.mkdir(exist_ok=True)
    for goarch,arch in [('arm64','aarch64_cortex-a53'),('amd64','x86_64')]:
        binary=ROOT/'work'/f'sing-box-{goarch}';plain=ROOT/'work'/f'sing-box-{goarch}.plain'
        if not binary.is_file() or not plain.is_file(): raise ValueError('Build binaries first')
        # APK reads real filesystem permissions. DrvFS may expose every file as
        # 0777 even after chmod; stage on the native Linux temporary filesystem.
        staging=tempfile.TemporaryDirectory(prefix='sing-box-x-payload-')
        payload=Path(staging.name)/arch
        shutil.copytree(ROOT/'packaging/files',payload)
        (payload/'usr/bin').mkdir(parents=True);shutil.copyfile(binary,payload/'usr/bin/sing-box')
        (payload/'etc/sing-box').mkdir();(payload/'usr/share/sing-box-x').mkdir(parents=True)
        info={**c,'architecture':arch,'binary_bytes':binary.stat().st_size,'plain_binary_bytes':plain.stat().st_size,'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest()}
        (payload/'usr/share/sing-box-x/build.json').write_text(json.dumps(info,indent=2)+'\n')
        files=[]
        for f in sorted(payload.rglob('*')):
            if f.is_file():
                name=f.relative_to(payload).as_posix();mode=0o755 if name in ('usr/bin/sing-box','etc/init.d/sing-box') else 0o644
                f.chmod(mode);os.utime(f,(int(os.environ.get('SOURCE_DATE_EPOCH','0')),)*2)
                if stat.S_IMODE(f.stat().st_mode)!=mode: raise ValueError('Payload filesystem cannot preserve package permissions')
                files.append(('./'+name,f.read_bytes(),mode))
        size=sum(len(data) for _,data,_ in files)
        control=f"Package: sing-box-x\nVersion: {v}-{rev}\nArchitecture: {arch}\nInstalled-Size: {size}\nDepends: ca-bundle, kmod-tun\nConflicts: sing-box, sing-box-tiny, sing-box-extended\nProvides: sing-box\nSection: net\nLicense: GPL-3.0-or-later\nMaintainer: Screamshow\nSource: https://github.com/Screamshow/sing-box-x\nDescription: Compact sing-box for Forkop X with uTLS and Clash API, packed with UPX\n"
        controls=[('./control',control.encode(),0o644),('./conffiles',b'/etc/config/sing-box\n',0o644)]
        package=tar_bytes([('debian-binary',b'2.0\n',0o644),('control.tar.gz',tar_bytes(controls),0o644),('data.tar.gz',tar_bytes(files),0o644)])
        (out/f'sing-box-x_{v}-{rev}_{arch}.ipk').write_bytes(package)
        meta=payload/'lib/apk/packages';meta.mkdir(parents=True)
        (meta/'sing-box-x.list').write_text(''.join(name[1:]+'\n' for name,_,_ in files))
        for suffix in ('conffiles','conffiles_static'):
            (meta/f'sing-box-x.{suffix}').write_text('/etc/config/sing-box\n')
        for f in meta.iterdir(): f.chmod(0o644);os.utime(f,(int(os.environ.get('SOURCE_DATE_EPOCH','0')),)*2)
        for directory in [payload,*[f for f in payload.rglob('*') if f.is_dir()]]:
            directory.chmod(0o755);os.utime(directory,(int(os.environ.get('SOURCE_DATE_EPOCH','0')),)*2)
            if stat.S_IMODE(directory.stat().st_mode)!=0o755: raise ValueError('Payload directory permissions are not 0755')
        cmd=[a.apk_tool,'mkpkg','--files',str(payload),'--output',str(out/f'sing-box-x_{v}-r{rev}_{arch}.apk')]
        fields={'name':'sing-box-x','version':f'{v}-r{rev}','arch':arch,'description':'Compact sing-box for Forkop X with uTLS and Clash API, packed with UPX','license':'GPL-3.0-or-later','origin':'sing-box-x','maintainer':'Screamshow','url':'https://github.com/Screamshow/sing-box-x','depends':'ca-bundle kmod-tun !sing-box !sing-box-tiny !sing-box-extended','provides':'sing-box'}
        for key,value in fields.items():cmd+=['-I',f'{key}:{value}']
        if os.geteuid()==0:
            for f in [payload,*payload.rglob('*')]:os.chown(f,0,0)
            subprocess.run(cmd,check=True)
        else:
            subprocess.run(['unshare','-r','sh','-c','chown -R 0:0 "$1"; shift; exec "$@"','sh',str(payload),*cmd],check=True)
        (out/f'sing-box-x-{v}-linux-{goarch}-upx.tar.gz').write_bytes(tar_bytes([('sing-box',binary.read_bytes(),0o755)]))
        staging.cleanup()
if __name__=='__main__':main()
