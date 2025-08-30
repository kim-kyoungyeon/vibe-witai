# 윗미(Wit.me) - 유튜브 팬미팅 모임 플랫폼 PRD

## 1. 프로젝트 개요
- **목표:** 유튜브 영상 댓글 감성분석을 기반으로 긍정적인 팬들이 모여 팬미팅을 진행할 수 있는 모임 플랫폼
- **핵심 가치:**
  - 감성분석 기반 모임(긍정 댓글 팬만 참여)
  - 간편한 모임 생성(유튜브 URL만 입력)
  - 팬 커뮤니티/네트워킹

## 2. 기술 스택
- **백엔드:** Go, Firebase Auth, Firebase Firestore, YouTube Data API, OpenAI API
- **프론트엔드:** Go Templates, HTML/CSS/JS

## 3. 핵심 기능 (MVP)
### 1단계: 회원가입/로그인 (Firebase Auth)
- 이메일/비밀번호 회원가입, 로그인, 로그아웃

### 2단계: 모임 생성 (유튜브 URL → 감성분석 → 모임 생성)
- 유튜브 URL 입력 → 댓글 감성분석(긍정/부정/중립)
- 긍정 댓글 작성자만 모임 참여 가능
- 모임 정보 입력/저장(모임명, 설명, 일시, 최대 인원)
- 한 사용자는 최대 5개 모임 생성 가능

### 3단계: 모임 목록/상세/참가 신청
- 전체 모임 목록/상세 페이지
- 긍정 댓글 작성자만 참가 신청 가능(이름/이메일 입력)
- 모임 생성자는 참가자 목록 확인, 승인/거절, CSV 다운로드
- 내가 생성/참여한 모임 목록

## 4. 데이터 구조 (Firestore)
- **Users**: uid, email, displayName, createdAt, meetingCount
- **Meetings**: meetingId, creatorId, youtubeUrl, videoTitle, videoChannel, meetingName, description, meetingDate, maxParticipants, createdAt, status
- **Participants**: participantId, meetingId, email, name, youtubeComment, status, appliedAt

## 5. 페이지 구조
- 메인(/): 모임 목록, 로그인/회원가입, 모임 생성
- 로그인(/login), 회원가입(/signup)
- 모임 생성(/create): 유튜브 URL, 감성분석 결과, 모임 정보 입력
- 모임 상세(/meeting/:id): 정보, 참가 신청, 참가자 목록(생성자)
- 모임 관리(/manage/:id): 참가자 승인/거절, 다운로드
- 내 모임(/my-meetings): 내가 생성/참여한 모임

## 6. 개발 단계별 요구사항
### Phase 1 (MVP)
1. Firebase Auth 연동: 이메일/비밀번호 회원가입 및 로그인
2. 유튜브 URL로 모임 생성(댓글 감성분석, 긍정 댓글 추출, 모임 정보 입력/저장)
3. 모임 목록/상세/참가 신청(긍정 댓글 작성자만)
4. 모임 생성자: 참가자 목록 확인, 승인/거절, CSV 다운로드
5. 내가 생성/참여한 모임 목록

### Phase 2 (확장)
- 팬 랭킹/등급, 어드민 대시보드, 유료 모임, 네트워킹, 알림 등

---

**이 파일은 Wit.me 프로젝트의 공식 PRD 및 단계별 요구사항 정의서입니다.**
