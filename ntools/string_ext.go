package ntools

import (
	"strings"
	"unicode"
)

type NString struct {
	S string
}

func (ns *NString) CutStr(start, end int) string {
	return ns.S[start:end]
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
func (ns *NString) UnderscoreToCamelcase(title bool) string {
	return lowerToCamelcase(ns.S, "_", title)
}

// for example: transfer BrowseBySet to browse_by_set
func (ns *NString) CamelcaseToUnderscore() string {
	return camelcaseToLower(ns.S, "_")
}

func lowerToCamelcase(str string, sp string, title bool) string {
	var method string
	sli := strings.Split(str, sp)
	for i, v := range sli {
		if i == 0 {
			if title {
				method += strings.Title(v)
			} else {
				method += v
			}
		} else {
			method += strings.Title(v)
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
