#!/usr/bin/env python3
"""Port the pinned XHTTP review series, retaining X's TLS and HTTP/1.1+2 scope."""
import argparse
import hashlib
import json
import subprocess
from pathlib import Path


def adapt(source: Path) -> None:
    # Keep the upstream patches intact for provenance; omit the HTTP/3 implementation
    # from the effective source. X never silently enables QUIC for XHTTP.
    for name in ('common/tls/quic_config.go', 'common/tls/utls_quic.go',
                 'transport/v2rayxhttp/http3.go', 'transport/v2rayxhttp/http3_test.go',
                 'transport/v2rayxhttp/reset_http3_test.go'):
        (source / name).unlink()
    path = source / 'transport/v2rayxhttp/http3_stub.go'
    text = path.read_text().replace('//go:build !with_quic\n\n', '')
    text = text.replace('needs a build with QUIC', 'is not supported by this HTTP/1.1+HTTP/2 X candidate')
    path.write_text(text, newline='\n')
    path = source / 'transport/v2rayxhttp/review_test.go'
    text = path.read_text()
    old = '\tfor name, test := range map[string]struct {'
    if text.count(old) != 1:
        raise ValueError('header-limit regression changed')
    # Default XMUX lazily opens three connections; this regression specifically
    # asserts preservation of one socket and must pin that policy.
    text = text.replace(old, '\toptions.Xmux = &option.V2RayXHTTPXmuxOptions{MaxConnections: xrange(1, 1)}\n' + old)
    path.write_text(text, newline='\n')
    path = source / 'transport/v2rayxhttp/client.go'
    text = path.read_text()
    old = '\treality := isReality(tlsConfig)\n\tversion := decideHTTPVersion(tlsConfig, reality)'
    if text.count(old) != 1:
        raise ValueError('XHTTP newLeg changed')
    text = text.replace(old, '''\tif tlsConfig != nil && len(tlsConfig.NextProtos()) == 1 && tlsConfig.NextProtos()[0] == "h3" {
\t\treturn "", nil, E.New("xhttp: HTTP/3 is not supported by this HTTP/1.1+HTTP/2 X candidate")
\t}
''' + old)
    path.write_text(text, newline='\n')
    # The first two regressions in this file test HTTP/2 but were behind with_quic.
    path = source / 'transport/v2rayxhttp/review2_test.go'
    text = path.read_text().split('func TestXHTTPHTTP3CloseDuringDial')[0]
    text = text.split('// Closing the client while a QUIC connection')[0]
    text = text.replace('//go:build with_quic\n\n', '')
    text = text.replace('\t"github.com/sagernet/sing/common/json/badoption"\n', '')
    path.write_text(text, newline='\n')
    path = source / 'go.mod'
    text = path.read_text()
    old = '\tgithub.com/klauspost/cpuid/v2 v2.3.0 // indirect\n'
    if text.count(old) != 1:
        raise ValueError('cpuid dependency changed; review before porting')
    text = text.replace(old, '').replace('\tgithub.com/keybase/go-keychain v0.0.1\n',
        '\tgithub.com/keybase/go-keychain v0.0.1\n\tgithub.com/klauspost/cpuid/v2 v2.3.0\n')
    path.write_text(text, newline='\n')
    # Adapt 0068: advertise only features actually ported, no urltest/decode-link.
    path = source / 'cmd/sing-box/cmd_version.go'
    text = path.read_text().replace('\tos.Stdout.WriteString(version)',
        '\tversion += "Features: transport.xhttp,transport.xhttp.http1,transport.xhttp.http2\\n"\n\tos.Stdout.WriteString(version)')
    path.write_text(text, newline='\n')
    helper = Path(__file__).with_name('xhttp-http2-pool.go')
    (source / 'transport/v2rayxhttp/settings_pool.go').write_bytes(helper.read_bytes())
    path = source / 'transport/v2rayxhttp/pool.go'
    text = path.read_text()
    old = '\treturn connection\n}\n\nfunc (c *http2Connection) roundTrip'
    if text.count(old) != 1:
        raise ValueError('HTTP/2 connection constructor changed')
    text = text.replace(old, '''\tconnection.transport.ConnPool = &settingsConnPool{
\t\ttransport: connection.transport,
\t\tdial: func(ctx context.Context) (net.Conn, error) {
\t\t\treturn connection.transport.DialTLSContext(ctx, "tcp", serverAddr.String(), nil)
\t\t},
\t}
''' + old)
    path.write_text(text, newline='\n')


def prepare(source: Path) -> None:
    patches = Path(__file__).resolve().parents[1] / 'patches/xhttp'
    manifest = json.loads((patches / 'manifest.json').read_text())
    manifest['adapter_sha256'] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    manifest['http2_pool_sha256'] = hashlib.sha256(Path(__file__).with_name('xhttp-http2-pool.go').read_bytes()).hexdigest()
    tests = Path(__file__).with_name('xhttp-tests')
    manifest['tests'] = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(tests.glob('*.go'))}
    marker = source / '.forkop-xhttp.json'
    if marker.exists():
        if json.loads(marker.read_text()) != manifest:
            raise ValueError('XHTTP manifest changed; use a clean source checkout')
        return
    for entry in manifest['patches']:
        patch = patches / entry['file']
        if hashlib.sha256(patch.read_bytes()).hexdigest() != entry['sha256']:
            raise ValueError(f'XHTTP patch digest mismatch: {patch.name}')
        if patch.name.startswith('0068-'):
            continue
        args = ['git', 'apply']
        if patch.name.startswith('0031-'):
            args += ['--include=common/readwait/*']
        # Podkop go.mod has unrelated forks; promote the same pinned cpuid below.
        if patch.name.startswith('0046-'):
            args += ['--exclude=go.mod']
        subprocess.run(args + ['--check', str(patch)], cwd=source, check=True)
        subprocess.run(args + [str(patch)], cwd=source, check=True)
    adapt(source)
    for test in tests.glob('*.go'):
        destination = source / 'transport/v2rayxhttp' / test.name
        if destination.exists():
            raise ValueError(f'refusing to overwrite {destination}')
        destination.write_bytes(test.read_bytes())
    subprocess.run(['gofmt', '-w', 'transport/v2rayxhttp', 'option/v2ray_xhttp.go',
                    'cmd/sing-box/cmd_version.go'], cwd=source, check=True)
    manifest['effective_files'] = {
        str(p.relative_to(source)): hashlib.sha256(p.read_bytes()).hexdigest()
        for p in sorted((source / 'transport/v2rayxhttp').glob('*.go'))}
    # Effective hashes are evidence, not an input to idempotency comparison.
    marker.write_text(json.dumps({k: v for k, v in manifest.items() if k != 'effective_files'}, indent=2)+'\n')
    (source / '.forkop-xhttp-effective.json').write_text(json.dumps(manifest, indent=2)+'\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source', type=Path)
    prepare(parser.parse_args().source.resolve())
