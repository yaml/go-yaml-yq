package yqlib

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const escape = "\x1b"

type colorAttribute int

const (
	colorReset       colorAttribute = 0
	colorFgGreen     colorAttribute = 32
	colorFgCyan      colorAttribute = 36
	colorFgHiBlack   colorAttribute = 90
	colorFgHiYellow  colorAttribute = 93
	colorFgHiMagenta colorAttribute = 95
)

func format(attr colorAttribute) string {
	return fmt.Sprintf("%s[%dm", escape, attr)
}

func colorizeAndPrint(data []byte, writer io.Writer) error {
	out := colorize(data)
	_, err := writer.Write(out)
	return err
}

func colorize(data []byte) []byte {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return data
	}
	switch trimmed[0] {
	case '{', '[':
		return colorizeJSON(data)
	default:
		return colorizeYAML(data)
	}
}

func colorizeJSON(data []byte) []byte {
	var out strings.Builder
	s := string(data)
	for i := 0; i < len(s); {
		switch ch := s[i]; {
		case ch == '"':
			end := scanJSONString(s, i)
			token := s[i:end]
			if isJSONObjectKey(s, end) {
				out.WriteString(wrap(colorFgCyan, token))
			} else {
				out.WriteString(wrap(colorFgGreen, token))
			}
			i = end
		case ch == '-' || ('0' <= ch && ch <= '9'):
			end := scanJSONNumber(s, i)
			out.WriteString(wrap(colorFgHiMagenta, s[i:end]))
			i = end
		case strings.HasPrefix(s[i:], "true") && isJSONBoundary(s, i+4):
			out.WriteString(wrap(colorFgHiMagenta, "true"))
			i += 4
		case strings.HasPrefix(s[i:], "false") && isJSONBoundary(s, i+5):
			out.WriteString(wrap(colorFgHiMagenta, "false"))
			i += 5
		case strings.HasPrefix(s[i:], "null") && isJSONBoundary(s, i+4):
			out.WriteString(wrap(colorFgHiMagenta, "null"))
			i += 4
		default:
			out.WriteByte(ch)
			i++
		}
	}
	return []byte(out.String())
}

func colorizeYAML(data []byte) []byte {
	var out strings.Builder
	lines := strings.SplitAfter(string(data), "\n")
	for _, line := range lines {
		lineEnding := ""
		body := line
		if strings.HasSuffix(body, "\n") {
			lineEnding = "\n"
			body = strings.TrimSuffix(body, "\n")
		}
		if strings.HasSuffix(body, "\r") {
			lineEnding = "\r" + lineEnding
			body = strings.TrimSuffix(body, "\r")
		}
		trimmed := strings.TrimSpace(body)
		if trimmed == "" {
			out.WriteString(body)
			out.WriteString(lineEnding)
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			out.WriteString(wrap(colorFgHiBlack, body))
			out.WriteString(lineEnding)
			continue
		}
		colored, _ := colorizeYAMLLine(body)
		out.WriteString(colored)
		out.WriteString(lineEnding)
	}
	return []byte(out.String())
}

func colorizeYAMLLine(line string) (string, bool) {
	commentAt := yamlCommentIndex(line)
	comment := ""
	body := line
	if commentAt >= 0 {
		body = line[:commentAt]
		comment = line[commentAt:]
	}

	colored := colorizeYAMLBody(body)
	if comment != "" {
		colored += wrap(colorFgHiBlack, comment)
	}
	return colored, yamlStartsBlockScalar(body)
}

func colorizeYAMLBody(body string) string {
	colonAt := yamlColonIndex(body)
	if colonAt >= 0 {
		return colorYAMLKeyPrefix(body[:colonAt]) + body[colonAt:colonAt+1] + colorYAMLValue(body[colonAt+1:])
	}
	trimmed := strings.TrimLeft(body, " ")
	prefixLen := len(body) - len(trimmed)
	if strings.HasPrefix(trimmed, "- ") {
		return body[:prefixLen] + "- " + colorYAMLValue(trimmed[2:])
	}
	return body[:prefixLen] + colorYAMLValue(trimmed)
}

func colorYAMLKeyPrefix(prefix string) string {
	trimmed := strings.TrimRight(prefix, " ")
	return wrap(colorFgCyan, trimmed) + prefix[len(trimmed):]
}

func colorYAMLValue(value string) string {
	var out strings.Builder
	for i := 0; i < len(value); {
		ch := value[i]
		if ch == ' ' || ch == '\t' || strings.ContainsRune("[]{}:,", rune(ch)) {
			out.WriteByte(ch)
			i++
			continue
		}
		if ch == '"' || ch == '\'' {
			end := scanYAMLQuoted(value, i)
			out.WriteString(wrap(colorFgGreen, value[i:end]))
			i = end
			continue
		}
		if ch == '&' || ch == '*' {
			end := scanYAMLWord(value, i+1)
			out.WriteString(wrap(colorFgHiYellow, value[i:end]))
			i = end
			continue
		}
		end := scanYAMLWord(value, i)
		token := value[i:end]
		switch {
		case isYAMLNumber(token), isYAMLBoolOrNull(token):
			out.WriteString(wrap(colorFgHiMagenta, token))
		default:
			out.WriteString(wrap(colorFgGreen, token))
		}
		i = end
	}
	return out.String()
}

func wrap(attr colorAttribute, s string) string {
	if s == "" {
		return s
	}
	return format(attr) + s + format(colorReset)
}

func scanJSONString(s string, start int) int {
	for i := start + 1; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == '"' {
			return i + 1
		}
	}
	return len(s)
}

func isJSONObjectKey(s string, afterString int) bool {
	for i := afterString; i < len(s); i++ {
		if unicode.IsSpace(rune(s[i])) {
			continue
		}
		return s[i] == ':'
	}
	return false
}

func scanJSONNumber(s string, start int) int {
	i := start
	if i < len(s) && s[i] == '-' {
		i++
	}
	for i < len(s) && '0' <= s[i] && s[i] <= '9' {
		i++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && '0' <= s[i] && s[i] <= '9' {
			i++
		}
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		for i < len(s) && '0' <= s[i] && s[i] <= '9' {
			i++
		}
	}
	return i
}

func isJSONBoundary(s string, i int) bool {
	if i >= len(s) {
		return true
	}
	return unicode.IsSpace(rune(s[i])) || strings.ContainsRune(",]}", rune(s[i]))
}

func yamlCommentIndex(s string) int {
	quote := byte(0)
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if quote == '"' && ch == '\\' && !escaped {
				escaped = true
				continue
			}
			if ch == quote && !escaped {
				quote = 0
			}
			escaped = false
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			continue
		}
		if ch == '#' && (i == 0 || unicode.IsSpace(rune(s[i-1]))) {
			return i
		}
	}
	return -1
}

func yamlColonIndex(s string) int {
	quote := byte(0)
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if quote == '"' && ch == '\\' && !escaped {
				escaped = true
				continue
			}
			if ch == quote && !escaped {
				quote = 0
			}
			escaped = false
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			continue
		}
		if ch == ':' && (i+1 == len(s) || unicode.IsSpace(rune(s[i+1]))) {
			return i
		}
	}
	return -1
}

func yamlStartsBlockScalar(body string) bool {
	trimmed := strings.TrimSpace(body)
	return strings.HasSuffix(trimmed, ": |") ||
		strings.HasSuffix(trimmed, ": >") ||
		strings.HasSuffix(trimmed, ": |-") ||
		strings.HasSuffix(trimmed, ": >-") ||
		strings.HasSuffix(trimmed, ": |+") ||
		strings.HasSuffix(trimmed, ": >+")
}

func scanYAMLQuoted(s string, start int) int {
	quote := s[start]
	for i := start + 1; i < len(s); i++ {
		if quote == '"' && s[i] == '\\' {
			i++
			continue
		}
		if s[i] == quote {
			return i + 1
		}
	}
	return len(s)
}

func scanYAMLWord(s string, start int) int {
	i := start
	for i < len(s) {
		if unicode.IsSpace(rune(s[i])) || strings.ContainsRune("[]{}:,", rune(s[i])) {
			break
		}
		i++
	}
	return i
}

func isYAMLBoolOrNull(s string) bool {
	switch strings.ToLower(s) {
	case "true", "false", "null", "~":
		return true
	default:
		return false
	}
}

func isYAMLNumber(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[i] == '-' || s[i] == '+' {
		i++
	}
	digit := false
	for i < len(s) && '0' <= s[i] && s[i] <= '9' {
		digit = true
		i++
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && '0' <= s[i] && s[i] <= '9' {
			digit = true
			i++
		}
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '-' || s[i] == '+') {
			i++
		}
		expDigit := false
		for i < len(s) && '0' <= s[i] && s[i] <= '9' {
			expDigit = true
			i++
		}
		digit = digit && expDigit
	}
	return digit && i == len(s)
}
