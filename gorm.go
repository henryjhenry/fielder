package main

import (
	"strings"
)

type gormTagParser struct{}

func (*gormTagParser) Parse(tag string) string {
	tag = strings.Trim(tag, "`")
	tl := len(tag)
	pl := 4 // len("gorm")
	for i := 0; i < tl; i++ {
		if i+pl > tl {
			return ""
		}
		if string(tag[i:i+pl]) != "gorm" {
			continue
		}
		// `gorm:"axxx"`, j is index('a')
		j := i + pl + 2

		nl := 6 // len("column")
		for ; j < tl; j++ {
			if j+nl > tl {
				return ""
			}
			if string(tag[j:j+nl]) != "column" {
				continue
			}
			k := j + nl + 1
			s := k
			for ; k < tl; k++ {
				if tag[k] == '"' || tag[k] == ';' {
					return tag[s:k]
				}
			}
		}
	}
	return ""
}
