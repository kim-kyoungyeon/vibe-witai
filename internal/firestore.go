package internal

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var firestoreClient *firestore.Client

func InitFirestore(projectID string) error {
	ctx := context.Background()
	sa := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(sa))
	if err != nil {
		return err
	}
	firestoreClient = client
	return nil
}

// Meeting 구조체 예시
type Meeting struct {
	MeetingID       string `firestore:"meetingId"`
	CreatorID       string `firestore:"creatorId"`
	YoutubeUrl      string `firestore:"youtubeUrl"`
	VideoTitle      string `firestore:"videoTitle"`
	VideoChannel    string `firestore:"videoChannel"`
	MeetingName     string `firestore:"meetingName"`
	Description     string `firestore:"description"`
	MeetingDate     string `firestore:"meetingDate"`
	CreatedAt       int64  `firestore:"createdAt"`
	MaxParticipants int    `firestore:"maxParticipants"`
	Status          string `firestore:"status"`
}

// 모임 생성
func CreateMeeting(ctx context.Context, m Meeting) error {
	_, _, err := firestoreClient.Collection("meetings").Add(ctx, m)
	return err
}

// 모임 목록 조회
func GetMeetings(ctx context.Context) ([]Meeting, error) {
	docs, err := firestoreClient.Collection("meetings").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	meetings := make([]Meeting, 0, len(docs))
	for _, doc := range docs {
		var m Meeting
		doc.DataTo(&m)
		meetings = append(meetings, m)
	}
	return meetings, nil
}

// 모임 상세 조회
func GetMeetingByID(ctx context.Context, meetingID string) (Meeting, error) {
	docs, err := firestoreClient.Collection("meetings").Where("meetingId", "==", meetingID).Documents(ctx).GetAll()
	if err != nil || len(docs) == 0 {
		return Meeting{}, err
	}
	var m Meeting
	docs[0].DataTo(&m)
	return m, nil
}

type Participant struct {
	Name  string `firestore:"name"`
	Email string `firestore:"email"`
}

// 참가자 목록 조회
func GetParticipants(ctx context.Context, meetingID string) ([]Participant, error) {
	docs, err := firestoreClient.Collection("participants").Where("meetingId", "==", meetingID).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	participants := make([]Participant, 0, len(docs))
	for _, doc := range docs {
		var p Participant
		doc.DataTo(&p)
		participants = append(participants, p)
	}
	return participants, nil
}
