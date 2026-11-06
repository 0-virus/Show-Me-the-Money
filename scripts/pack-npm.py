"""Create an npm release tarball with portable Unix executable permissions."""
import io
import shutil
import subprocess
import sys
import tarfile

from package import ROOT, DIST, VERSION


def main():
    npm = shutil.which('npm')
    if not npm:
        raise RuntimeError('npm is required to build the package')
    DIST.mkdir(parents=True, exist_ok=True)
    subprocess.run([npm, 'pack', '--pack-destination', str(DIST)], cwd=ROOT, check=True)
    archive = DIST / f'halmasy1-moneycopy-{VERSION}.tgz'
    staged = archive.with_suffix('.tmp')
    # npm pack on Windows records native executables as 0644. Normalize the
    # release artifact so admin-owned/read-only installs also work on macOS.
    with tarfile.open(archive, 'r:gz') as source, tarfile.open(staged, 'w:gz', format=tarfile.PAX_FORMAT) as target:
        for member in source.getmembers():
            if not member.isfile():
                raise ValueError(f'Unexpected non-file entry: {member.name}')
            content = source.extractfile(member).read()
            member.mode = 0o755 if member.name in (
                'package/bin/moneycopy.cjs', 'package/native/darwin-universal/moneycopy',
            ) else 0o644
            target.addfile(member, io.BytesIO(content))
    staged.replace(archive)
    subprocess.run([sys.executable, str(ROOT / 'scripts/verify-npm.py'), str(archive)], cwd=ROOT, check=True)
    print(f'Publish this verified artifact: {archive}')


if __name__ == '__main__':
    main()
