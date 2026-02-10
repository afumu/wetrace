package wordcloud

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// WordCloudResult 词云结果
type WordCloudResult struct {
	TotalMessages int         `json:"total_messages"`
	TotalWords    int         `json:"total_words"`
	Words         []*WordItem `json:"words"`
}

// WordItem 词频项
type WordItem struct {
	Text  string `json:"text"`
	Count int    `json:"count"`
}

// 中文停用词表
var stopWords = map[string]bool{
	"的": true, "了": true, "是": true, "在": true, "我": true,
	"你": true, "他": true, "她": true, "它": true, "们": true,
	"这": true, "那": true, "有": true, "和": true, "就": true,
	"不": true, "也": true, "都": true, "要": true, "会": true,
	"可以": true, "没有": true, "什么": true, "一个": true, "我们": true,
	"自己": true, "他们": true, "没": true, "很": true, "到": true,
	"说": true, "对": true, "吗": true, "啊": true, "呢": true,
	"吧": true, "嗯": true, "哦": true, "哈": true, "呀": true,
	"嘛": true, "哎": true, "唉": true, "喔": true, "噢": true,
	"把": true, "被": true, "让": true, "给": true, "从": true,
	"去": true, "来": true, "上": true, "下": true, "里": true,
	"中": true, "大": true, "小": true, "多": true, "少": true,
	"个": true, "人": true, "还": true, "能": true, "做": true,
	"看": true, "想": true, "知道": true, "时候": true, "现在": true,
	"因为": true, "所以": true, "但是": true, "如果": true, "这个": true,
	"那个": true, "已经": true, "可能": true, "应该": true, "怎么": true,
	"为什么": true, "这样": true, "那样": true, "一下": true, "一些": true,
	"然后": true, "或者": true, "而且": true, "虽然": true, "不过": true,
	"只是": true, "其实": true, "觉得": true, "比较": true, "一样": true,
}

// Analyze 对文本列表进行词频统计，返回词云结果
func Analyze(texts []string, limit int) *WordCloudResult {
	if limit <= 0 {
		limit = 100
	}

	freq := make(map[string]int)
	totalWords := 0

	for _, text := range texts {
		words := tokenize(text)
		for _, w := range words {
			if !stopWords[w] {
				freq[w]++
				totalWords++
			}
		}
	}

	// 转为切片并排序
	items := make([]*WordItem, 0, len(freq))
	for text, count := range freq {
		items = append(items, &WordItem{Text: text, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return &WordCloudResult{
		TotalMessages: len(texts),
		TotalWords:    totalWords,
		Words:         items,
	}
}

// tokenize 对文本进行简单分词
// 使用二元组（bigram）方式提取中文词汇，同时提取英文单词
func tokenize(text string) []string {
	var words []string
	var chineseRunes []rune
	var englishWord strings.Builder

	for _, r := range text {
		if isChinese(r) {
			// 如果之前有英文单词，先保存
			if englishWord.Len() > 0 {
				w := strings.ToLower(englishWord.String())
				if utf8.RuneCountInString(w) >= 2 {
					words = append(words, w)
				}
				englishWord.Reset()
			}
			chineseRunes = append(chineseRunes, r)
		} else {
			// 处理累积的中文字符，提取二元组
			words = append(words, extractBigrams(chineseRunes)...)
			chineseRunes = chineseRunes[:0]

			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				englishWord.WriteRune(r)
			} else {
				if englishWord.Len() > 0 {
					w := strings.ToLower(englishWord.String())
					if utf8.RuneCountInString(w) >= 2 {
						words = append(words, w)
					}
					englishWord.Reset()
				}
			}
		}
	}

	// 处理末尾
	words = append(words, extractBigrams(chineseRunes)...)
	if englishWord.Len() > 0 {
		w := strings.ToLower(englishWord.String())
		if utf8.RuneCountInString(w) >= 2 {
			words = append(words, w)
		}
	}

	return words
}

// extractBigrams 从中文字符序列中提取二元组
func extractBigrams(runes []rune) []string {
	if len(runes) < 2 {
		return nil
	}
	var bigrams []string
	for i := 0; i < len(runes)-1; i++ {
		bigrams = append(bigrams, string(runes[i:i+2]))
	}
	return bigrams
}

// isChinese 判断是否为中文字符
func isChinese(r rune) bool {
	return unicode.Is(unicode.Han, r)
}
