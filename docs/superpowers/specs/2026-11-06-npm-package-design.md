# npm 배포 구성

사용자 요청: 앞서 설명한 Go 네이티브 실행 파일 + Node 실행기 구조를 npm 사용자명 `halmasy1`으로 구성한다.

- 패키지: `@halmasy1/moneycopy`, 버전 `1.0.3`, 명령 `moneycopy`.
- Windows x64와 macOS Intel/Apple Silicon 바이너리를 한 패키지에 포함한다. Node는 실행기이며 Go 설치는 불필요하다.
- 대안인 플랫폼별 optionalDependencies는 현재 파일 크기에 비해 배포할 패키지가 늘어나므로 사용하지 않는다. GitHub에서 설치 중 다운로드하는 방식도 추가 네트워크 의존성이 생겨 사용하지 않는다.
- root package.json의 bin/files로 실행기와 필요한 배포 파일만 포함한다. 네이티브 파일은 기존 버전별 dist에서 생성하며 Git에는 넣지 않는다.
- 실행기는 OS/아키텍처를 확인하고 인수·터미널 입출력·종료 상태를 전달한다. Mac 실행 권한이 빠진 압축 해제도 실행 시 복구한다.
- npm pack 전에 바이너리 버전과 배포 파일을 검증한다. pack:release는 최종 tarball의 Mac/CLI 실행 권한을 0755로 보존한다. 디렉토리 직접 publish는 재압축으로 권한이 사라질 수 있어 막고 검증된 tarball 게시 명령을 안내한다. 설치 시 Python/Go 또는 install 스크립트 실행은 필요하지 않다.
- 범위는 로컬 패키징과 설치·실행 검증이다. npm 로그인 및 공개 publish는 수행하지 않는다.
- 라이선스는 임의로 오픈소스 권한을 부여하지 않고 UNLICENSED로 명시한다. 포함 의존성 라이선스는 유지한다.
- Windows 실제 설치/실행, Node 테스트, tarball 포함 파일과 Mac 실행 파일 구조를 검증한다. Mac 실기기 실행은 미검증으로 표시한다.
