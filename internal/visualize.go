package internal

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// CountWords: 댓글 배열에서 단어 빈도 맵 생성
func CountWords(comments []string) map[string]int {
	freq := make(map[string]int)
	wordRe := regexp.MustCompile(`\p{L}+[\p{L}\p{N}']*`)
	for _, text := range comments {
		words := wordRe.FindAllString(strings.ToLower(text), -1)
		for _, w := range words {
			if len(w) < 2 { // 한 글자 단어 제외
				continue
			}
			freq[w]++
		}
	}
	return freq
}

// TopNWords: 빈도 맵에서 상위 N개 단어 추출 (키워드 인사이트용)
func TopNWords(freq map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range freq {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool { return ss[i].Value > ss[j].Value })
	res := make([]string, 0, n)
	for i := 0; i < n && i < len(ss); i++ {
		res = append(res, ss[i].Key)
	}
	return res
}

// GeneratePieChart: 감성분석 결과 비율 파이차트 SVG 생성
func GeneratePieChart(labels []string, filePath string) error {
	count := map[string]int{}
	for _, l := range labels {
		count[l]++
	}
	items := make([]opts.PieData, 0, len(count))
	for k, v := range count {
		items = append(items, opts.PieData{Name: k, Value: v})
	}
	pie := charts.NewPie()
	pie.AddSeries("감성분석", items)
	pie.SetGlobalOptions()
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return pie.Render(f)
}

func GenerateWordCloud(words map[string]int, filePath string) error {
	// 워드클라우드 SVG 생성 예시 (실제 구현 필요)
	wc := charts.NewWordCloud()
	items := make([]opts.WordCloudData, 0, len(words))
	for k, v := range words {
		items = append(items, opts.WordCloudData{Name: k, Value: v})
	}
	wc.AddSeries("wordcloud", items)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return wc.Render(f)
}
