"""Package native executables; preserve Unix executable bits in macOS ZIP."""
import hashlib
import json
import os
from pathlib import Path
import plistlib
import shutil
import struct
import subprocess
import zipfile

ROOT = Path(__file__).resolve().parents[1]
VERSION = '1.0.3'
DIST = ROOT / 'dist' / f'v{VERSION}'
BUILD = ROOT / 'artifacts/build'


def universal_binary(parts):
    alignment = 16384
    offset = alignment
    headers, chunks = [], []
    for path in parts:
        data = path.read_bytes()
        magic, cpu, subtype = struct.unpack_from('<III', data)
        if magic != 0xfeedfacf:
            raise ValueError(f'Not a 64-bit Mach-O: {path}')
        headers.append(struct.pack('>IIIII', cpu, subtype, offset, len(data), 14))
        chunks.append((offset, data))
        offset = (offset + len(data) + alignment - 1) // alignment * alignment
    result = bytearray(struct.pack('>II', 0xcafebabe, len(parts)) + b''.join(headers))
    for offset, data in chunks:
        result.extend(b'\0' * (offset-len(result)))
        result.extend(data)
    return bytes(result)


def text(path, content):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding='utf-8', newline='\n')


def zip_folder(folder):
    dest = DIST / f'MoneyCopy-{folder.name}-v{VERSION}.zip'
    with zipfile.ZipFile(dest, 'w', zipfile.ZIP_DEFLATED, compresslevel=9) as archive:
        for path in sorted(folder.rglob('*')):
            if not path.is_file():
                continue
            executable = path.name in ('MoneyCopy', 'moneycopy', 'Play.command')
            info = zipfile.ZipInfo(path.relative_to(DIST).as_posix())
            info.create_system = 3
            info.external_attr = (0o100755 if executable else 0o100644) << 16
            info.compress_type = zipfile.ZIP_DEFLATED
            archive.writestr(info, path.read_bytes())
    return dest


def third_party_notices():
    go = ROOT / '.tools/go/bin/go.exe'
    exe = str(go) if go.exists() else 'go'
    # Only linked production modules, rather than unrelated dependency test tools.
    template = '{{if .Module}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}'
    raw = subprocess.check_output([exe, 'list', '-deps', '-f', template, './cmd/moneycopy'], cwd=ROOT, text=True)
    modules = []
    for row in sorted(set(raw.splitlines())):
        if not row or row.startswith('moneycopy|'):
            continue
        name, version, directory = row.split('|')
        modules.append(dict(Path=name, Version=version, Dir=directory))
    notices = ['THIRD PARTY LICENSES\n']
    for module in modules:
        notices.append(f"\n{'='*70}\n{module['Path']} {module['Version']}\n")
        directory = Path(module['Dir'])
        licenses = [p for p in directory.iterdir() if p.is_file() and p.name.upper().startswith(('LICENSE', 'COPYING'))]
        if not licenses:
            raise ValueError(f"Missing license: {module['Path']}")
        for path in licenses:
            notices.append(path.read_text(encoding='utf-8'))
    goroot = subprocess.check_output([exe, 'env', 'GOROOT'], text=True).strip()
    notices.append('\nGo runtime\n' + (Path(goroot) / 'LICENSE').read_text())
    return '\n'.join(notices)


def main():
    windows, mac = DIST/'Windows', DIST/'macOS'
    windows.mkdir(parents=True, exist_ok=True)
    mac.mkdir(parents=True, exist_ok=True)
    shutil.copy2(BUILD/'MoneyCopy.exe', windows/'MoneyCopy.exe')
    bundle = mac/'MoneyCopy.app'/'Contents'
    resources = bundle/'Resources'
    resources.mkdir(parents=True, exist_ok=True)
    (resources/'moneycopy').write_bytes(universal_binary([BUILD/'moneycopy-amd64', BUILD/'moneycopy-arm64']))
    text(bundle/'MacOS'/'MoneyCopy', '#!/bin/sh\nset -eu\nHERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)\nexec /usr/bin/open -a Terminal "$HERE/../Resources/Play.command"\n')
    text(resources/'Play.command', '#!/bin/sh\nset -eu\nHERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)\nexport LANG="${LANG:-en_US.UTF-8}"\nexport COLORTERM=truecolor\nexec "$HERE/moneycopy"\n')
    plist = dict(CFBundleExecutable='MoneyCopy', CFBundleIdentifier='local.moneycopy.game',
                 CFBundleName='Money Copy', CFBundleDisplayName='가만히 있어도 돈이 복사가 되면 얼마나 좋을까',
                 CFBundlePackageType='APPL', CFBundleShortVersionString=VERSION, CFBundleVersion='5',
                 LSMinimumSystemVersion='13.0', NSHighResolutionCapable=True)
    (bundle/'Info.plist').write_bytes(plistlib.dumps(plist))
    notices = third_party_notices()
    for folder in (windows, mac):
        shutil.copy2(ROOT/'README.md', folder/'README.md')
        (folder/'docs').mkdir(exist_ok=True)
        shutil.copy2(ROOT/'docs/verification.md', folder/'docs/verification.md')
        text(folder/'THIRD-PARTY-LICENSES.txt', notices)
    archives = [zip_folder(windows),zip_folder(mac)]
    text(DIST/'SHA256SUMS.txt', ''.join(f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n' for p in archives))
    for path in archives:
        print(f'{path} ({path.stat().st_size:,} bytes)')


if __name__ == '__main__':
    main()
