#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const { constants } = require('node:os');

function resolveBinary(platform, arch, root) {
  if (platform === 'win32' && arch === 'x64') {
    return path.join(root, 'native/windows-x64/MoneyCopy.exe');
  }
  if (platform === 'darwin' && (arch === 'x64' || arch === 'arm64')) {
    return path.join(root, 'native/darwin-universal/moneycopy');
  }
  throw new Error(`지원하지 않는 환경입니다: ${platform}/${arch}. Windows x64 또는 macOS Intel/Apple Silicon에서 실행해 주세요.`);
}

function run(args, options = {}) {
  const {
    platform = process.platform,
    arch = process.arch,
    root = path.join(__dirname, '..'),
    spawn = spawnSync,
    error = console.error,
  } = options;
  try {
    const binary = resolveBinary(platform, arch, root);
    if (!fs.existsSync(binary)) {
      throw new Error('게임 실행 파일이 없습니다. @halmasy1/moneycopy 패키지를 다시 설치해 주세요.');
    }
    // Windows-created archives may omit Unix execute bits. No install hook is
    // required: restore only the packaged game's permission when it is launched.
    if (platform === 'darwin') {
      try {
        fs.accessSync(binary, fs.constants.X_OK);
      } catch {
        fs.chmodSync(binary, fs.statSync(binary).mode | 0o111);
      }
    }
    const child = spawn(binary, args, { stdio: 'inherit', shell: false });
    if (child.error) {
      throw new Error(`게임을 실행하지 못했습니다 (${child.error.code || 'ERROR'}): ${child.error.message}`);
    }
    if (child.signal) {
      return 128 + (constants.signals[child.signal] || 1);
    }
    return child.status ?? 1;
  } catch (err) {
    error(`MoneyCopy: ${err.message}`);
    return 1;
  }
}

if (require.main === module) {
  process.exitCode = run(process.argv.slice(2));
}

module.exports = { resolveBinary, run };
