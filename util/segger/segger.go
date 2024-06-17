package segger

import (
	"langchaingo-learn/util"
	"langchaingo-learn/util/goahocorasick"
	"regexp"
	"strconv"
	"strings"

	"log"

	"github.com/yanyiwu/gojieba"
)

var (
	segmenter        *gojieba.Jieba
	delimterReg      *regexp.Regexp
	blankReg         *regexp.Regexp
	engReg           *regexp.Regexp
	digitalReg       *regexp.Regexp
	punc             = make(map[string]bool, 100)
	multiWordMatcher *goahocorasick.Machine
)

const (
	use_hmm = true
)

func init() {
	delimterReg = regexp.MustCompile(`[,\|\t]`)
	blankReg = regexp.MustCompile(`\s+`)
	engReg = regexp.MustCompile(`[a-zA-Z0-9\.\+#_]+`)
	digitalReg = regexp.MustCompile(`\d+`)
	segmenter = gojieba.NewJieba()
	lines := util.ReadAllLines(util.ProjectRootPath + "data/punc.txt")
	for _, ele := range lines {
		if len(ele) > 0 {
			punc[ele] = true
		}
	}
}

// 从文件中载入词典，用于分词。这些词包括用户自定义词典、停用词、关键词、地名
func LoadUserDict() {
	const DEFAULT_FREQ = 20000 //默认词频
	const DEFAULT_TAG = ""     //默认词性
	DictFiles := []string{
		util.ProjectRootPath + "data/stop.txt",
		util.ProjectRootPath + "data/user_dict.txt",
		util.ProjectRootPath + "data/country.txt",
		util.ProjectRootPath + "data/province.txt",
		util.ProjectRootPath + "data/city.txt"}
	cnt := 0
	for _, dictFile := range DictFiles {
		lines := util.ReadAllLines(dictFile)
		for _, line := range lines {
			words := delimterReg.Split(strings.TrimSpace(strings.ToLower(line)), -1)
			for _, word := range words {
				word = strings.TrimSpace(word)
				if len([]rune(word)) > 1 && !digitalReg.MatchString(word) { //排除纯数字
					segmenter.AddWordEx(word, DEFAULT_FREQ, DEFAULT_TAG)
					cnt++
				}
			}
		}
	}

	lines := util.ReadAllLines(util.ProjectRootPath + "data/keywords.txt")
	dict := [][]rune{}
	for _, line := range lines {
		arr := delimterReg.Split(strings.TrimSpace(strings.ToLower(line)), -1) //注意：word freq tag 3列用\t分隔，不要用空格
		word := arr[0]
		if strings.Contains(arr[0], " ") {
			dict = append(dict, []rune(arr[0]))
			word = strings.Replace(arr[0], " ", "_", -1)
		}
		freq := DEFAULT_FREQ
		tag := DEFAULT_TAG
		if len(arr) >= 2 {
			if v, err := strconv.Atoi(arr[1]); err == nil {
				freq = v
			} else {
				log.Printf("invalid user dict word frequency: %s", arr[1])
			}
		}
		if len(arr) >= 3 {
			tag = arr[2]
		}
		segmenter.AddWordEx(word, freq, tag)
	}

	multiWordMatcher = new(goahocorasick.Machine)
	if err := multiWordMatcher.Build(dict); err != nil {
		panic(err)
	}

	log.Printf("load %d word dict from %v", cnt, DictFiles)
}

// 多个英文单词构成一个词的情况，用"_"连起来
func JoinMultiWord(sentence string) string {
	sentRune := []rune(sentence)
	terms := multiWordMatcher.MultiPatternSearch(sentRune, false)
	for _, term := range terms {
		beginPos := term.Pos
		for i, r := range term.Word {
			if r == ' ' {
				sentRune[beginPos+i] = '_'
			}
		}
	}
	return string(sentRune)
}

// Seg 分词
func Seg(sentence string) []string {
	sentence = JoinMultiWord(sentence)
	words := Cut(sentence)
	// fmt.Println(words)

	//把中英文标点去掉
	rect := make([]string, 0, len(words))
	for _, ele := range words {
		ele = strings.TrimSpace(ele)
		if len(ele) == 0 {
			continue
		}
		if _, exists := punc[ele]; exists {
			continue
		}
		if _, exists := punc[ele[0:1]]; exists && ele[0:1] != "." { //如果word第一个字符是标点，则把这个标点删掉
			ele = ele[1:]
		}
		if len(ele) > 1 && (ele[len(ele)-1:] == "-" || ele[len(ele)-1:] == "." || (ele[len(ele)-1:] == "+" && ele[len(ele)-2:len(ele)-1] != "+")) { //如果word最后一个字符是.或-或+，则把最后这个字符删掉
			ele = ele[:len(ele)-1]
		}
		// ele = strings.Replace(ele, "_", " ", -1)
		rect = append(rect, ele)
	}
	return rect
}

// Cut 分词
func Cut(sentence string) []string {
	rect := []string{}
	parts := _split(sentence)
	for _, part := range parts {
		if engReg.MatchString(part) {
			rect = append(rect, part)
		} else {
			rect = append(rect, segmenter.Cut(part, use_hmm)...)
		}
	}
	return rect
}

func _split(sentence string) []string {
	//遇到空格，强行分开
	parts := blankReg.Split(sentence, -1)
	if len(parts) > 0 {
		rect := []string{}
		for _, s := range parts {
			rect = append(rect, _split_by_eng(s)...)
		}
		return rect
	}
	return []string{}
}

func _split_by_eng(sentence string) []string {
	rect := []string{}
	ms := engReg.FindAllStringSubmatchIndex(sentence, -1)
	if len(ms) > 0 {
		begin := 0
		for i := 0; i < len(ms); i++ {
			if ms[i][0] > begin {
				s := strings.TrimSpace(sentence[begin:ms[i][0]])
				if s != "" {
					rect = append(rect, s)
				}
			}
			s := strings.TrimSpace(sentence[ms[i][0]:ms[i][1]])
			if s != "" {
				rect = append(rect, s)
			}
			begin = ms[i][1]
		}
		rect = append(rect, sentence[begin:])
	} else {
		rect = append(rect, sentence)
	}
	return rect
}

func FreeSegger() {
	segmenter.Free()
}
