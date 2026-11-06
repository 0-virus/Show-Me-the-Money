"""Stage verified native releases for npm pack; no downloads or publication."""
import json
import plistlib
import re
import shutil
import subprocess
import sys

from package import ROOT, DIST, VERSION


def main():
    manifest = json.loads((ROOT / 'package.json').read_text(encoding='utf-8'))
    go_main = (ROOT / 'cmd/moneycopy/main.go').read_text(encoding='utf-8')
    go_version = re.search(r'const version = "([^"]+)"', go_main)
    if manifest['version'] != VERSION or not go_version or go_version[1] != VERSION:
        raise ValueError('package.json, Go, and scripts/package.py versions must match')
    if not DIST.is_dir():
        raise FileNotFoundError(f'Build native releases first: scripts/build.ps1 (missing {DIST})')

    subprocess.run([sys.executable, str(ROOT / 'scripts/verify-packages.py')], cwd=ROOT, check=True)
    app = DIST / 'macOS/MoneyCopy.app/Contents'
    plist = plistlib.loads((app / 'Info.plist').read_bytes())
    if plist['CFBundleShortVersionString'] != VERSION:
        raise ValueError('Mac bundle version does not match npm package')
    sources = {
        'windows-x64/MoneyCopy.exe': DIST / 'Windows/MoneyCopy.exe',
        'darwin-universal/moneycopy': app / 'Resources/moneycopy',
    }
    host_source = sources.get('windows-x64/MoneyCopy.exe' if sys.platform == 'win32' else 'darwin-universal/moneycopy')
    if sys.platform in ('win32', 'darwin'):
        version = subprocess.check_output([str(host_source), '--version'], text=True).strip()
        if version != f'Money Copy {VERSION}':
            raise ValueError(f'Stale native executable: {version}')
    for relative, source in sources.items():
        target = ROOT / 'native' / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target)
        target.chmod(0o755)
    shutil.copyfile(DIST / 'Windows/THIRD-PARTY-LICENSES.txt', ROOT / 'THIRD-PARTY-LICENSES.txt')
    (ROOT / 'bin/moneycopy.cjs').chmod(0o755)
    print(f'Ready to pack {manifest["name"]}@{VERSION}')


if __name__ == '__main__':
    main()
