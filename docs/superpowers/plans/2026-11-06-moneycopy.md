# Moneycopy Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 실제 터미널에서 돈을 적립하고 Windows/Mac 배포 파일을 제공한다.
**Architecture:** 계산·시간 규칙은 독립 game 패키지, 입력·도트 렌더링은 terminal 패키지, main은 실제 화면과 이벤트 루프를 연결한다.
**Tech Stack:** Go, tcell v2, Python 표준 라이브러리(배포 ZIP/Mach-O 결합).
**Spec:** ../specs/2026-11-06-moneycopy-design.md

## Global Constraints
- 제목: 가만히 있어도 돈이 복사가 되면 얼마나 좋을까
- 월 180시간, 연 12개월. 5초마다 5초치 시급 적립. 30분치 시급 성과급.
- 실제 터미널, 네트워크 없이 실행, OS별 배포 디렉토리, 굵직한 작업마다 커밋.

## Review Focus
- 소수 시급의 누적 및 큰 금액 오버플로 → 정수 분수 계산 테스트.
- 18시 경계·절전·시계 역행 → 마감 시각과 중복 적립 테스트.
- 성과급 중복 클릭·착지 후 클릭 → 이벤트 및 모델 테스트.
- 작은 창·한글 폭·리사이즈 → 시뮬레이션 화면 테스트.
- Mac 권한·아키텍처, Windows 독립 실행 → 패키지 검사 및 실행 파일 검증.

## Task 1: 규칙
Files: internal/game/game.go, internal/game/game_test.go, go.mod
Interface: New(salary int64, annual bool, now time.Time) (*Session,error), Advance(now), Finish(now), ClaimBonus(now), Base(), Bonus(), Total(), Hourly(), Monthly().
- [ ] 월/연봉 환산, 5초 경계, 지연된 여러 틱, 마감 경계, 중복 보너스 실패 테스트 작성 및 실패 확인.
- [ ] 정수 기반 모델 구현, 전체 테스트 통과, 커밋.

## Task 2: 터미널 게임
Files: internal/terminal/app.go, render.go, app_test.go, cmd/moneycopy/main.go
Interface: New(screen tcell.Screen, now func() time.Time, random *rand.Rand), Handle(event), Tick(), Draw(), Run().
- [ ] 초기→게임→결과→초기, 입력 오류, 클릭, 리사이즈 시뮬레이션 테스트 및 실패 확인.
- [ ] 입력, 상태바, 낙하 지폐, 지폐더미, 성과급 히트박스 구현.
- [ ] 테스트·빌드 및 실제 EXE 플레이, 커밋.

## Task 3: 배포
Files: scripts/build.ps1, scripts/package.py, README.md, docs/verification.md
- [ ] Windows amd64/arm64, Mac amd64/arm64 교차 빌드.
- [ ] Universal Mach-O, 앱 실행기, OS별 디렉토리/ZIP 생성.
- [ ] 패키지 헤더·권한·무결성, 전체 테스트/vet, 배포 EXE 확인.
- [ ] 설명서와 검증 결과 기록, 독립 리뷰·필요 수정, 커밋.
