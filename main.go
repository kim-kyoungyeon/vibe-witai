package main

import (
	"log"
	"net/http"
	"os"

	"youtube-analyzer/internal"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	// 환경변수 체크
	if os.Getenv("YOUTUBE_API_KEY") == "" || os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("환경변수 YOUTUBE_API_KEY, OPENAI_API_KEY를 설정하세요.")
	}
	// Firestore 초기화
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("환경변수 FIREBASE_PROJECT_ID(Firebase 프로젝트 ID)를 설정하세요.")
	}
	if err := internal.InitFirestore(projectID); err != nil {
		log.Fatal("Firestore 초기화 실패: ", err)
	}
	// Firebase Admin SDK 초기화 불필요 (REST API만 사용)

	http.HandleFunc("/signup", internal.SignupHandler)
	http.HandleFunc("/login", internal.LoginHandler)
	http.HandleFunc("/logout", internal.LogoutHandler)

	http.HandleFunc("/", internal.IndexHandler)
	http.HandleFunc("/analyze", internal.AuthRequired(internal.AnalyzeHandler))
	http.HandleFunc("/create", internal.AuthRequired(internal.CreateMeetingHandler))
	http.HandleFunc("/my-meetings", internal.AuthRequired(internal.MyMeetingsHandler))
	http.HandleFunc("/meeting", internal.AuthRequired(internal.MeetingDetailHandler))
	http.HandleFunc("/manage", internal.AuthRequired(internal.ManageMeetingHandler))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("서버 시작: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
