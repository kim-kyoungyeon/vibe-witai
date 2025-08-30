package internal

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestCreateAndGetMeeting(t *testing.T) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		t.Skip("FIREBASE_PROJECT_ID 환경변수 필요")
	}
	if err := InitFirestore(projectID); err != nil {
		t.Fatalf("Firestore 초기화 실패: %v", err)
	}
	m := Meeting{
		MeetingName:     "테스트 모임",
		YoutubeUrl:      "https://youtube.com/test",
		Description:     "테스트 설명",
		MeetingDate:     time.Now().Format(time.RFC3339),
		MaxParticipants: 10,
		CreatedAt:       time.Now().Unix(),
		Status:          "active",
	}
	err := CreateMeeting(context.Background(), m)
	if err != nil {
		t.Fatalf("모임 생성 실패: %v", err)
	}
	meetings, err := GetMeetings(context.Background())
	if err != nil {
		t.Fatalf("모임 목록 조회 실패: %v", err)
	}
	found := false
	for _, mt := range meetings {
		if mt.MeetingName == "테스트 모임" {
			found = true
			break
		}
	}
	if !found {
		t.Error("생성한 모임이 목록에 없음")
	}
}
