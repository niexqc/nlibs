package ntools

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type NString struct {
	S string
}

func StrFromGbkBytes(bytes []byte) *NString {
	utf8Data, _, _ := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), bytes)
	return &NString{S: string(utf8Data)}
}

func (ns *NString) CutStr(start, end int) string {
	length := len(ns.S)
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > length {
		start = length
	}
	if end > length {
		end = length
	}
	if start > end {
		return ""
	}
	return ns.S[start:end]
}

var blankRegexp = regexp.MustCompile(`\s+`)

func (ns *NString) ReplaceAllBlank(toStr string) string {
	return blankRegexp.ReplaceAllString(ns.S, toStr)
}

// CutString ...
func (ns *NString) CutString(length int) string {
	if len(ns.S) > length {
		if length > 6 {
			resultStr := ns.S[0 : length-3]
			return resultStr + "..."
		}
		return ns.S[0:length]
	}
	return ns.S
}

// for example: transfer browse_by_set to BrowseBySet
func (ns *NString) Under2Camel(title bool) string {
	return lowerToCamelcase(ns.S, "_", title)
}

// for example: transfer BrowseBySet to browse_by_set
func (ns *NString) Camel2Under() string {
	return camelcaseToLower(ns.S, "_")
}

// titleCase 将字符串首字母大写，其余字母保持原样。
// strings.Title 已被标记为 deprecated（对 Unicode 处理有缺陷且会改小写部分），这里手工实现。
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func lowerToCamelcase(str string, sp string, title bool) string {
	var method string
	sli := strings.Split(str, sp)
	for i, v := range sli {
		if i == 0 {
			if title {
				method += titleCase(v)
			} else {
				method += v
			}
		} else {
			method += titleCase(v)
		}
	}
	return method
}

func camelcaseToLower(str string, sp string) string {
	return strings.Join(camelcaseToSlice(str, true, -1), sp)
}

func camelcaseToSlice(str string, toLower bool, limit int) []string {
	var words []string
	l := 0
	i := 1

	for s := str; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l < 1 || (limit > 0 && limit == i) {
			l = len(s)
		}

		if toLower {
			words = append(words, strings.ToLower(s[:l]))
		} else {
			words = append(words, s[:l])
		}

		i++
	}

	return words
}
