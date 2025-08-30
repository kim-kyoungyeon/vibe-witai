package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	meetings, _ := GetMeetings(r.Context())
	tmpl, _ := template.ParseFiles("web/templates/index.html")
	tmpl.Execute(w, map[string]interface{}{
		"Meetings": meetings,
	})
}

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var videoID string
	if r.FormValue("random") == "1" {
		id, err := FetchRandomPopularVideoID()
		if err != nil {
			http.Error(w, "인기 영상 조회 실패", 500)
			return
		}
		videoID = id
	} else {
		input := r.FormValue("video_id")
		videoID = ParseVideoID(input)
		if videoID == "" {
			http.Error(w, "유효한 YouTube 영상 ID 또는 URL을 입력하세요.", 400)
			return
		}
	}

	meta, _ := FetchVideoMeta(videoID)

	comments, err := FetchComments(videoID)
	if err != nil {
		http.Error(w, "유튜브 댓글 수집 실패", 500)
		return
	}
	// 1. 댓글 100개로 제한 (랜덤 샘플링)
	maxComments := 100
	if len(comments) > maxComments {
		rand.Seed(time.Now().UnixNano())
		perm := rand.Perm(len(comments))[:maxComments]
		sampled := make([]Comment, 0, maxComments)
		for _, idx := range perm {
			sampled = append(sampled, comments[idx])
		}
		comments = sampled
	}
	// 댓글 텍스트 배열
	commentTexts := make([]string, 0, len(comments))
	for _, c := range comments {
		commentTexts = append(commentTexts, c.Text)
	}
	// 2. 감성분석 병렬 처리 (최대 5개 동시)
	type resultItem struct {
		idx   int
		label string
	}
	results := make([]string, len(comments))
	ch := make(chan resultItem, len(comments))
	sem := make(chan struct{}, 5) // 동시 5개 제한
	var wg sync.WaitGroup
	for i, c := range comments {
		wg.Add(1)
		go func(idx int, text string) {
			defer wg.Done()
			sem <- struct{}{}
			label, _ := AnalyzeSentiment(text)
			ch <- resultItem{idx, label}
			<-sem
		}(i, c.Text)
	}
	wg.Wait()
	close(ch)
	for r := range ch {
		results[r.idx] = r.label
	}
	// 워드클라우드/차트 파일 경로
	wordcloudPath := "web/static/wordcloud.html"
	piechartPath := "web/static/piechart.html"
	// 워드클라우드 생성
	freq := CountWords(commentTexts)
	_ = GenerateWordCloud(freq, wordcloudPath)
	// 파이차트 생성
	_ = GeneratePieChart(results, piechartPath)
	// 상위 키워드 추출 및 인사이트 요약
	topKeywords := TopNWords(freq, 5)
	insight, _ := SummarizeKeywords(topKeywords)
	// 감성분석 결과 개수/비율 계산
	pos, neg, neu := 0, 0, 0
	total := len(results)
	for _, label := range results {
		switch label {
		case "긍정":
			pos++
		case "부정":
			neg++
		case "중립":
			neu++
		}
	}
	percent := func(n int) int {
		if total == 0 {
			return 0
		}
		return n * 100 / total
	}
	// 결과 템플릿 렌더링
	tmpl, err := template.ParseFiles("web/templates/result.html")
	if err != nil {
		http.Error(w, "템플릿 에러", 500)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Comments":      comments,
		"Sentiments":    results,
		"WordCloudPath": "/static/wordcloud.html",
		"PieChartPath":  "/static/piechart.html",
		"Insight":       insight,
		"VideoTitle":    meta.Title,
		"VideoChannel":  meta.Channel,
		"VideoThumb":    meta.Thumbnail,
		"PosCount":      pos,
		"NegCount":      neg,
		"NeuCount":      neu,
		"PosPercent":    percent(pos),
		"NegPercent":    percent(neg),
		"NeuPercent":    percent(neu),
		"TotalCount":    total,
	})
}

// 회원가입 핸들러
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("web/templates/signup.html")
		tmpl.Execute(w, nil)
		return
	}
	// POST
	email := r.FormValue("email")
	password := r.FormValue("password")
	displayName := r.FormValue("displayName")
	apiKey := os.Getenv("FIREBASE_WEB_API_KEY")
	url := "https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=" + apiKey
	body := map[string]interface{}{
		"email":             email,
		"password":          password,
		"displayName":       displayName,
		"returnSecureToken": true,
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "회원가입 실패: "+err.Error(), 400)
		return
	}
	defer resp.Body.Close()
	var result struct {
		IDToken string `json:"idToken"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "회원가입 실패: "+err.Error(), 400)
		return
	}
	if result.IDToken == "" {
		http.Error(w, "회원가입 실패: "+result.Error.Message, 400)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// 로그인 핸들러
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("web/templates/login.html")
		tmpl.Execute(w, nil)
		return
	}
	// POST
	email := r.FormValue("email")
	password := r.FormValue("password")
	// Firebase Auth REST API로 로그인(토큰 발급)
	idToken, err := FirebaseEmailPasswordLogin(email, password)
	if err != nil {
		http.Error(w, "로그인 실패: "+err.Error(), 400)
		return
	}
	// 쿠키에 토큰 저장(간단 세션)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    idToken,
		Path:     "/",
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// 로그아웃 핸들러
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Firebase REST API로 이메일/비밀번호 로그인(토큰 발급)
func FirebaseEmailPasswordLogin(email, password string) (string, error) {
	apiKey := os.Getenv("FIREBASE_WEB_API_KEY")
	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + apiKey
	body := map[string]interface{}{
		"email":             email,
		"password":          password,
		"returnSecureToken": true,
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		IDToken string `json:"idToken"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.IDToken == "" {
		return "", fmt.Errorf("로그인 실패: %s", result.Error.Message)
	}
	return result.IDToken, nil
}

// 인증 체크 함수
func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return false
	}
	// 실제 토큰 검증은 생략(추후 확장)
	return true
}

// 인증 미들웨어
func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// 모임 생성 페이지/처리
func CreateMeetingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("web/templates/create.html")
		tmpl.Execute(w, nil)
		return
	}
	// POST
	youtubeUrl := r.FormValue("youtubeUrl")
	meetingName := r.FormValue("meetingName")
	description := r.FormValue("description")
	meetingDate := r.FormValue("meetingDate")
	maxParticipants := r.FormValue("maxParticipants")
	if meetingName == "" {
		comments, _ := FetchComments(ParseVideoID(youtubeUrl))
		pos, neg, neu := 0, 0, 0
		for _, c := range comments {
			label, _ := AnalyzeSentiment(c.Text)
			switch label {
			case "긍정":
				pos++
			case "부정":
				neg++
			case "중립":
				neu++
			}
		}
		summary := "긍정: " + itoa(pos) + ", 부정: " + itoa(neg) + ", 중립: " + itoa(neu) + " (상위 20개 댓글 기준)"
		tmpl, _ := template.ParseFiles("web/templates/create.html")
		tmpl.Execute(w, map[string]interface{}{
			"ShowForm":        true,
			"AnalysisSummary": summary,
			"YoutubeUrl":      youtubeUrl,
		})
		return
	}
	// Firestore에 모임 정보 저장
	ctx := r.Context()
	m := Meeting{
		YoutubeUrl:      youtubeUrl,
		MeetingName:     meetingName,
		Description:     description,
		MeetingDate:     meetingDate,
		MaxParticipants: atoi(maxParticipants),
		CreatedAt:       time.Now().Unix(),
		Status:          "active",
	}
	err := CreateMeeting(ctx, m)
	if err != nil {
		http.Error(w, "모임 저장 실패: "+err.Error(), 500)
		return
	}
	w.Write([]byte("모임 생성 완료!"))
}

func itoa(i int) string { return fmt.Sprintf("%d", i) }

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// 내 모임 목록
func MyMeetingsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: 내가 생성/참여한 모임 목록 조회
	w.Write([]byte("내 모임 목록 (구현 예정)"))
}

// 모임 상세 페이지
func MeetingDetailHandler(w http.ResponseWriter, r *http.Request) {
	meetingID := r.URL.Query().Get("id")
	if meetingID == "" {
		http.Error(w, "잘못된 접근", 400)
		return
	}
	ctx := r.Context()
	if r.Method == http.MethodPost {
		// 참가자 신청 저장(스켈레톤)
		name := r.FormValue("name")
		email := r.FormValue("email")
		// TODO: 참가자 Firestore 저장
		_ = name
		_ = email
		// 리다이렉트
		http.Redirect(w, r, "/meeting?id="+meetingID, http.StatusSeeOther)
		return
	}
	meeting, _ := GetMeetingByID(ctx, meetingID)
	participants, _ := GetParticipants(ctx, meetingID)
	tmpl, _ := template.ParseFiles("web/templates/meeting.html")
	tmpl.Execute(w, map[string]interface{}{
		"Meeting":      meeting,
		"Participants": participants,
	})
}

// 모임 관리(참가자 승인/거절, 다운로드 등)
func ManageMeetingHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: 참가자 관리, 승인/거절, CSV 다운로드
	w.Write([]byte("모임 관리 (구현 예정)"))
}
