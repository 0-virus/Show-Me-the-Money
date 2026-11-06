"""Validate the distributable bytes, executable metadata, and archive integrity."""
import hashlib
from pathlib import Path
import plistlib
import struct
import zipfile

root = Path(__file__).resolve().parents[1]
dist = root / 'dist'

def require(condition, message):
    if not condition:
        raise AssertionError(message)

exe = (dist/'Windows/MoneyCopy.exe').read_bytes()
require(exe[:2] == b'MZ', 'Missing Windows executable')
pe = struct.unpack_from('<I', exe, 60)[0]
require(exe[pe:pe+4] == b'PE\0\0', 'Bad PE header')
require(struct.unpack_from('<H',exe,pe+4)[0] == 0x8664, 'Not Windows x64')
require(struct.unpack_from('<H',exe,pe+24+68)[0] == 3, 'Not console subsystem')

app = dist/'macOS/MoneyCopy.app/Contents'
data = (app/'Resources/moneycopy').read_bytes()
require(struct.unpack_from('>II',data) == (0xcafebabe,2), 'Not universal Mach-O')
cpus = []
for i in range(2):
    cpu, sub, offset, length, align = struct.unpack_from('>IIIII',data,8+i*20)
    require(offset % (1 << align) == 0 and offset+length <= len(data), 'Bad slice bounds')
    require(struct.unpack_from('<III',data,offset) == (0xfeedfacf,cpu,sub), 'Slice mismatch')
    cpus.append(cpu)
require(set(cpus)=={0x01000007,0x0100000c}, 'Missing Intel or Apple Silicon')
plist = plistlib.loads((app/'Info.plist').read_bytes())
require((app/'MacOS'/plist['CFBundleExecutable']).is_file(), 'Missing app launcher')
for path in [app/'MacOS/MoneyCopy',app/'Resources/Play.command']:
    require(path.read_bytes().startswith(b'#!/bin/sh\n'), 'Invalid shell entry point')
    require(b'\r' not in path.read_bytes(), 'CRLF breaks Mac launcher')

for folder in ['Windows','macOS']:
    with zipfile.ZipFile(dist/f'MoneyCopy-{folder}.zip') as archive:
        require(archive.testzip() is None, 'Corrupt ZIP')
        for member in archive.infolist():
            require(archive.read(member) == (dist/member.filename).read_bytes(), 'Archive differs from folder')
            if Path(member.filename).name in ['moneycopy','MoneyCopy','Play.command']:
                require(member.create_system == 3 and (member.external_attr >> 16) & 0o111 == 0o111, 'Missing execute bits')
for line in (dist/'SHA256SUMS.txt').read_text().splitlines():
    digest,name = line.split()
    require(hashlib.sha256((dist/name).read_bytes()).hexdigest()==digest, 'Checksum mismatch')
print('PASS: Windows x64 console EXE; Mac Intel + ARM64; launchers; ZIP permissions; checksums')
