package main

// The default parser parses the first tag to get the field name.

type defaultTagParser struct{}

func (*defaultTagParser) Parse(tag string) string {
	tag = tag[1 : len(tag)-1] // trim "`"
	tl := len(tag)
	for i := 0; i < tl; i++ {
		if tag[i] == ':' {
			// skip `:"`
			s := i + 2
			for j := s; j < tl; j++ {
				if tag[j] == '"' {
					return tag[s:j]
				}
			}
		}
	}
	return ""
}
