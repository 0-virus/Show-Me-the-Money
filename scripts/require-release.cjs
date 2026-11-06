// Publishing a directory repacks it and can lose Mac execute bits on Windows.
const { version } = require('../package.json');
console.error('먼저 npm run pack:release 로 배포 파일을 만들어 주세요.');
console.error(`게시 명령: npm publish ./dist/v${version}/halmasy1-moneycopy-${version}.tgz --access public`);
process.exitCode = 1;
