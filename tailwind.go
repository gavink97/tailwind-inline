package tailwindinline

import (
	b "bytes"
	"fmt"
	"log/slog"
	s "strings"
)

type field struct {
	Type  string
	Class string
}

func getClasses(str string) []string {
	twcls := s.SplitAfter(s.ToLower(str), `class="`)

	var classes []string
	for i, v := range twcls {
		if i != 0 {
			split := s.Split(v, `"`)
			classes = append(classes, split[0])
		}
	}

	var set = make(map[string]bool)
	var class []string
	for _, v := range classes {
		fields := s.Fields(v)
		for _, x := range fields {
			if set[x] {
				slog.Debug(fmt.Sprintf("%s already exists in fields", x))
			} else {
				class = append(class, x)
				set[x] = true
			}
		}
	}

	return class
}

func assignField(class string) field {
	mediaPrefix := []string{"sm:", "md:", "lg:", "xl:", "2xl:", "max-sm:",
		"max-md:", "max-lg:", "max-xl:", "max-2xl:", "min-[", "max-[",
		"motion-safe:", "contrast-more:", "contrast-less:", "dark:",
		"forced-colors:", "landscape:", "motion-reduce:", "print:",
		"portrait:"}

	for range mediaPrefix {
		for _, prefix := range mediaPrefix {
			if s.HasPrefix(class, prefix) {
				return field{Type: "media", Class: class}
			}
		}
	}

	return field{Type: "inline", Class: class}
}

func transform(class string) string {
	if s.ContainsRune(class, ':') {
		idx := s.Index(class, ":")
		class = s.Replace(class, ":", `\:`, idx)
		class = fmt.Sprintf(".%s", class)
		return class
	}

	return class
}

// condense variables if unsupported
func getInlineStyles(styles []byte, field field) string {
	class := transform(field.Class)

	idx := b.Index(styles, []byte(class))
	if idx == -1 {
		slog.Error(fmt.Sprintf("Class not present in styles sheet %s", class))
		return ""
	}

	cut := b.FieldsFunc(styles[idx:], func(r rune) bool {
		if r == '{' || r == '}' {
			return true
		}
		return false
	})

	return string(cut[1])
}

func getMediaQueries(styles []byte, field field) string {
	class := transform(field.Class)

	idx := b.Index(styles, []byte(class))
	if idx == -1 {
		slog.Error(fmt.Sprintf("Class not present in styles sheet %s", class))
		return ""
	}

	midx := b.LastIndex(styles[:idx], []byte("@media"))
	if midx == -1 {
		slog.Error(fmt.Sprintf("Class not present in styles sheet %s", class))
		return ""
	}

	cut := b.FieldsFunc(styles[midx:], func(r rune) bool {
		if r == '@' || r == '}' {
			return true
		}
		return false
	})

	return fmt.Sprintf("@%s}\n}", string(cut[0]))
}

func validateClassByFirstIndex(tpl string, field field) int {
	class := field.Class
	clone := s.Clone(tpl)

	for range s.Count(clone, class) {
		index := s.Index(clone, class)
		if index == -1 {
			slog.Error(fmt.Sprintf("Class not present in styles sheet %s", class))
			return -1
		}

		prefixs := []string{" ", `"`}

		for range prefixs {
			for _, x := range prefixs {
				if s.LastIndex(clone[:index], x)+1 == index {
					return index
				}
			}
		}

		rcls := fmt.Sprintf("%sz", class[1:])
		clone = s.Replace(clone, class, rcls, 1)
	}
	return -1
}

func replaceWithInlineCSS(tpl, css string, field field) string {
	class := field.Class
	css = s.ReplaceAll(css, "\n  ", "")
	css = s.ReplaceAll(css, "\n", "")

	for range s.Count(tpl, class) {
		index := validateClassByFirstIndex(tpl, field)
		if index == -1 {
			slog.Debug(fmt.Sprintf("There are no more valid %s inside tpl", class))
			break
		}

		oi := s.LastIndex(tpl[:index], "<")
		if oi == -1 {
			slog.Error(fmt.Sprintf("Missing opening tag at %d", oi))
			return tpl
		}

		ci := s.Index(tpl[index:], ">")
		if ci == -1 {
			slog.Error(fmt.Sprintf("Missing closing tag at %d", ci))
			return tpl
		}

		ci = ci + index

		if !s.Contains(s.ToLower(tpl[oi:ci]), `style="`) {
			styleAttr := fmt.Sprintf(`style="%s"`, css)
			tpl = fmt.Sprintf("%s %s%s", tpl[:ci], styleAttr, tpl[ci:])
		} else {
			if !s.Contains(tpl[oi:ci], css) {
				var cut int
				if s.HasSuffix(tpl[:ci], `"`) {
					cut = ci - 1
				} else {
					cut = ci
				}
				tpl = fmt.Sprintf(`%s%s"%s`, tpl[:cut], css, tpl[ci:])
			}
		}
		tpl = fmt.Sprintf("%s%s", tpl[:index], tpl[index+len(class):])
	}

	return tpl
}

func writeMediaQueries(tpl string, css []string) string {
	if !s.Contains(s.ToLower(tpl), "<head>") {
		slog.Info("There is no head tag present in your template")
		headtag := "<head>\n</head>"
		tpl = fmt.Sprintf("%s%s", headtag, tpl)
	}

	hoi := s.Index(s.ToLower(tpl), "<head>")
	if hoi == -1 {
		slog.Error("There is no head tag present in your template")
		return tpl
	}

	hci := s.Index(s.ToLower(tpl), "</head>")
	if hci == -1 {
		slog.Error("There is no head closing tag present in your template")
		return tpl
	}

	if !s.Contains(s.ToLower(tpl[hoi:hci]), "<style>") {
		slog.Info("There is no style tag present in your templates head")
		styletag := "<style>\n</style>"
		tpl = fmt.Sprintf("%s%s%s", tpl[:hoi+6], styletag, tpl[hci:])
		hci = hci + len(styletag)
	}

	media := s.Join(css, "\n")

	sci := s.Index(s.ToLower(tpl[hoi:hci]), "</style>")
	if sci == -1 {
		slog.Error("There is no style tag present in your templates head")
		return tpl
	}

	tpl = fmt.Sprintf("%s\n%s\n%s", tpl[:sci], media, tpl[sci:])

	return tpl
}

func Convert(tpl string, styles []byte) string {
	fields := getClasses(tpl)

	var mediaQueries []string
	for _, field := range fields {
		field := assignField(field)
		if field.Type == "media" {
			css := getMediaQueries(styles, field)
			mediaQueries = append(mediaQueries, css)
		} else {
			css := getInlineStyles(styles, field)
			tpl = replaceWithInlineCSS(tpl, css, field)
		}
	}

	tpl = writeMediaQueries(tpl, mediaQueries)
	// clean up old class attributes / spacing issues

	return tpl
}

// create styles for img display block
// style="display:block"
