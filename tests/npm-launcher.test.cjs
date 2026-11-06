const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { resolveBinary, run } = require('../bin/moneycopy.cjs');

test('selects only supported native platforms', () => {
  const root = path.resolve('package with spaces');
  assert.equal(resolveBinary('win32', 'x64', root), path.join(root, 'native/windows-x64/MoneyCopy.exe'));
  for (const arch of ['x64', 'arm64']) {
    assert.equal(resolveBinary('darwin', arch, root), path.join(root, 'native/darwin-universal/moneycopy'));
  }
  for (const [os, arch] of [['linux', 'x64'], ['win32', 'arm64'], ['darwin', 'ia32']]) {
    assert.throws(() => resolveBinary(os, arch, root), /지원하지 않는/);
  }
});

function fixture(t) {
  const parent = path.resolve('artifacts/npm-test-fixtures');
  fs.mkdirSync(parent, { recursive: true });
  const root = fs.mkdtempSync(path.join(parent, 'package with spaces-'));
  const nativeDir = path.join(root, 'native/windows-x64');
  fs.mkdirSync(nativeDir, { recursive: true });
  fs.writeFileSync(path.join(nativeDir, 'MoneyCopy.exe'), 'fixture');
  t.after(() => {
    assert.equal(path.dirname(path.resolve(root)), parent);
    fs.rmSync(root, { recursive: true });
  });
  const errors = [];
  return { root, platform: 'win32', arch: 'x64', error: (s) => errors.push(s), errors };
}

test('forwards arguments literally and inherits the terminal without a shell', (t) => {
  const options = fixture(t);
  const args = ['--version', 'literal & $(text)', 'two words'];
  options.spawn = (file, argv, config) => {
    assert.equal(file, resolveBinary('win32', 'x64', options.root));
    assert.deepEqual(argv, args);
    assert.equal(config.stdio, 'inherit');
    assert.equal(config.shell, false);
    return { status: 7 };
  };
  assert.equal(run(args, options), 7);
  assert.deepEqual(options.errors, []);
});

test('reports missing and unlaunchable binaries instead of succeeding', (t) => {
  const options = fixture(t);
  const missing = { ...options, root: path.join(options.root, 'missing'), spawn: () => assert.fail('must not spawn') };
  assert.equal(run([], missing), 1);
  assert.match(options.errors.pop(), /다시 설치/);
  options.spawn = () => ({ error: Object.assign(new Error('blocked'), { code: 'EACCES' }) });
  assert.equal(run([], options), 1);
  assert.match(options.errors.pop(), /EACCES/);
});

test('preserves signal termination and rejects unsupported platforms', (t) => {
  const options = fixture(t);
  options.spawn = () => ({ status: null, signal: 'SIGINT' });
  assert.equal(run([], options), 130);
  options.platform = 'linux';
  options.spawn = () => assert.fail('must not spawn');
  assert.equal(run([], options), 1);
  assert.match(options.errors.pop(), /지원하지 않는/);
});
