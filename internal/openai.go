package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

var openAIModel = "gpt-4o" // 최신 모델 우선 적용, 실패시 gpt-3.5-turbo로 fallback

// AnalyzeSentiment: 댓글 텍스트를 OpenAI로 감성분석 ('긍정', '부정', '중립' 중 하나)
func AnalyzeSentiment(text string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	prompt := "다음 문장의 감성을 '긍정', '부정', '중립' 중 하나로만 답해줘. 문장: " + text
	body := map[string]interface{}{
		"model": openAIModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// gpt-4o 실패시 gpt-3.5-turbo로 재시도
		if openAIModel != "gpt-3.5-turbo" {
			openAIModel = "gpt-3.5-turbo"
			return AnalyzeSentiment(text)
		}
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) > 0 {
		label := strings.TrimSpace(result.Choices[0].Message.Content)
		label = strings.ReplaceAll(label, ".", "") // 마침표 제거
		if strings.Contains(label, "긍정") {
			return "긍정", nil
		} else if strings.Contains(label, "부정") {
			return "부정", nil
		} else if strings.Contains(label, "중립") {
			return "중립", nil
		}
		return label, nil // 혹시 다른 답변이 오면 그대로 반환
	}
	return "분석불가", nil
}

// SummarizeKeywords: 주요 키워드 배열을 받아 인사이트 및 여론 분석 생성 (OpenAI API 활용)
func SummarizeKeywords(keywords []string) (string, error) {
	if len(keywords) == 0 {
		return "주요 키워드가 없습니다.", nil
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	prompt := "다음 키워드들을 바탕으로 유튜브 댓글의 주요 인사이트와 여론(주요 토픽, 논쟁점, 긍/부정 분위기 등)을 2~3문장으로 요약해줘.\n- 키워드: " + strings.Join(keywords, ", ") + "\n- 결과는 자연스러운 한글 문장으로, 여론의 전체적 분위기와 논쟁점, 긍/부정/중립 비율 등도 포함해서 요약해줘."
	body := map[string]interface{}{
		"model": openAIModel,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// gpt-4o 실패시 gpt-3.5-turbo로 재시도
		if openAIModel != "gpt-3.5-turbo" {
			openAIModel = "gpt-3.5-turbo"
			return SummarizeKeywords(keywords)
		}
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) > 0 {
		return strings.TrimSpace(result.Choices[0].Message.Content), nil
	}
	return "주요 키워드: " + strings.Join(keywords, ", "), nil
}
