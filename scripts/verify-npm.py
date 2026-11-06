"""Validate the exact npm tarball, not just the working tree."""
import hashlib
import json
from pathlib import Path
import struct
import sys
import tarfile

from package import ROOT, DIST, VERSION


def require(condition, message):
    if not condition:
        raise AssertionError(message)


def main():
    archive = Path(sys.argv[1]) if len(sys.argv) > 1 else DIST / f'halmasy1-moneycopy-{VERSION}.tgz'
    expected = {
        'package/package.json', 'package/README.md', 'package/THIRD-PARTY-LICENSES.txt',
        'package/bin/moneycopy.cjs', 'package/native/windows-x64/MoneyCopy.exe',
        'package/native/darwin-universal/moneycopy',
    }
    with tarfile.open(archive, 'r:gz') as tar:
        members = tar.getmembers()
        require(len(members) == len(expected) and {m.name for m in members} == expected, 'Unexpected or missing package files')
        require(all(m.isfile() for m in members), 'Links or special files must not be published')
        for member in members:
            if member.name in ('package/bin/moneycopy.cjs', 'package/native/darwin-universal/moneycopy'):
                require(member.mode & 0o111 == 0o111, f'Executable permission missing: {member.name}')
        files = {m.name: tar.extractfile(m).read() for m in members}
    meta = json.loads(files['package/package.json'])
    require(meta['name'] == '@halmasy1/moneycopy' and meta['version'] == VERSION, 'Wrong package identity')
    require(meta['bin'] == {'moneycopy': 'bin/moneycopy.cjs'}, 'Wrong CLI entry')
    require(meta['publishConfig']['access'] == 'public', 'Public publishing configuration missing')
    require(not any(k in meta.get('scripts', {}) for k in ('preinstall', 'install', 'postinstall', 'prepare')), 'Consumer install must not run build hooks')
    require(not meta.get('dependencies') and not meta.get('optionalDependencies'), 'Unexpected runtime downloads')
    launcher = files['package/bin/moneycopy.cjs']
    require(launcher.startswith(b'#!/usr/bin/env node\n') and b'\r' not in launcher, 'Launcher requires LF shebang')
    for name, data in files.items():
        require(data == (ROOT / name.removeprefix('package/')).read_bytes(), f'Tarball differs from source: {name}')
    exe = files['package/native/windows-x64/MoneyCopy.exe']
    require(exe == (DIST / 'Windows/MoneyCopy.exe').read_bytes(), 'Windows binary differs from verified release')
    mac = files['package/native/darwin-universal/moneycopy']
    require(mac == (DIST / 'macOS/MoneyCopy.app/Contents/Resources/moneycopy').read_bytes(), 'Mac binary differs from verified release')
    require(struct.unpack_from('>II', mac) == (0xcafebabe, 2), 'Mac binary must be universal')
    print(f'PASS: {meta["name"]}@{VERSION}; six allowed files; exact release binaries; LF launcher; no install hooks')
    print(f'SHA256: {hashlib.sha256(archive.read_bytes()).hexdigest()}')


if __name__ == '__main__':
    main()
