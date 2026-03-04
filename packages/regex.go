// name: regex
// description: Regular expression utilities
// author: roturbot
// requires: regexp, strings

type Regex struct{}

func (Regex) match(pattern any, text any) bool {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)

	matched, err := regexp.MatchString(patternStr, textStr)
	if err != nil {
		return false
	}
	return matched
}

func (Regex) find(pattern any, text any) string {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)

	re := regexp.MustCompile(patternStr)
	matches := re.FindString(textStr)

	return matches
}

func (Regex) findAll(pattern any, text any) []any {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)

	re := regexp.MustCompile(patternStr)
	matches := re.FindAllString(textStr, -1)

	result := make([]any, len(matches))
	for i, match := range matches {
		result[i] = match
	}
	return result
}

func (Regex) findSubmatch(pattern any, text any) []any {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)

	re := regexp.MustCompile(patternStr)
	matches := re.FindStringSubmatch(textStr)

	if matches == nil {
		return []any{}
	}

	result := make([]any, len(matches))
	for i, match := range matches {
		result[i] = match
	}
	return result
}

func (Regex) replace(pattern any, text any, replacement any) string {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)
	replacementStr := OSLtoString(replacement)

	re := regexp.MustCompile(patternStr)
	return re.ReplaceAllString(textStr, replacementStr)
}

func (Regex) replaceFunc(pattern any, text any, fn any) string {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)

	re := regexp.MustCompile(patternStr)

	result := re.ReplaceAllStringFunc(textStr, func(match string) string {
		fnResult := OSLcallFunc(fn, nil, []any{match})
		return OSLtoString(fnResult)
	})

	return result
}

func (Regex) split(pattern any, text any) []any {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)

	re := regexp.MustCompile(patternStr)
	parts := re.Split(textStr, -1)

	result := make([]any, len(parts))
	for i, part := range parts {
		result[i] = part
	}
	return result
}

func (Regex) count(pattern any, text any) int {
	patternStr := OSLtoString(pattern)
	textStr := OSLtoString(text)

	re := regexp.MustCompile(patternStr)
	matches := re.FindAllString(textStr, -1)

	return len(matches)
}

func (Regex) test(pattern any) bool {
	patternStr := OSLtoString(pattern)
	_, err := regexp.Compile(patternStr)
	return err == nil
}

func (Regex) escape(text any) string {
	return regexp.QuoteMeta(OSLtoString(text))
}

func (Regex) isValidEmail(email any) bool {
	emailStr := OSLtoString(email)
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regex.match(pattern, emailStr)
}

func (Regex) isValidURL(url any) bool {
	urlStr := OSLtoString(url)
	pattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(?:/.*)?$`
	return regex.match(pattern, urlStr)
}

func (Regex) isValidIPv4(ip any) bool {
	ipStr := OSLtoString(ip)
	pattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	return regex.match(pattern, ipStr)
}

func (Regex) isValidIPv6(ip any) bool {
	ipStr := OSLtoString(ip)
	pattern := `^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|::|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`
	return regex.match(pattern, ipStr)
}

func (Regex) isValidPhone(phone any) bool {
	phoneStr := strings.ReplaceAll(OSLtoString(phone), " ", "")
	phoneStr = strings.ReplaceAll(phoneStr, "-", "")
	phoneStr = strings.ReplaceAll(phoneStr, "(", "")
	phoneStr = strings.ReplaceAll(phoneStr, ")", "")

	pattern := `^\+?1?\d{10,14}$`
	return regex.match(pattern, phoneStr)
}

func (Regex) extractEmail(text any) string {
	textStr := OSLtoString(text)
	pattern := `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`
	return regex.find(pattern, textStr)
}

func (Regex) extractEmails(text any) []any {
	textStr := OSLtoString(text)
	pattern := `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`
	return regex.findAll(pattern, textStr)
}

func (Regex) extractURLs(text any) []any {
	textStr := OSLtoString(text)
	pattern := `https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(?:/[^\s]*)?`
	return regex.findAll(pattern, textStr)
}

func (Regex) extractHashtags(text any) []any {
	textStr := OSLtoString(text)
	pattern := `#[A-Za-z0-9_]+`
	return regex.findAll(pattern, textStr)
}

func (Regex) extractMentions(text any) []any {
	textStr := OSLtoString(text)
	pattern := `@[A-Za-z0-9_]+`
	return regex.findAll(pattern, textStr)
}

func (Regex) extractNumbers(text any) []any {
	textStr := OSLtoString(text)
	pattern := `\d+`
	return regex.findAll(pattern, textStr)
}

func (Regex) extractWords(text any) []any {
	textStr := OSLtoString(text)
	pattern := `\b[A-Za-z]+\b`
	return regex.findAll(pattern, textStr)
}

func (Regex) isAlpha(text any) bool {
	textStr := OSLtoString(text)
	pattern := `^[A-Za-z]+$`
	return regex.match(pattern, textStr)
}

func (Regex) isAlphanumeric(text any) bool {
	textStr := OSLtoString(text)
	pattern := `^[A-Za-z0-9]+$`
	return regex.match(pattern, textStr)
}

func (Regex) isNumeric(text any) bool {
	textStr := OSLtoString(text)
	pattern := `^\d+$`
	return regex.match(pattern, textStr)
}

func (Regex) isHexadecimal(text any) bool {
	textStr := OSLtoString(text)
	pattern := `^0x[0-9A-Fa-f]+$|^[0-9A-Fa-f]+$`
	return regex.match(pattern, textStr)
}

func (Regex) isBase64(text any) bool {
	textStr := strings.ReplaceAll(OSLtoString(text), "\n", "")
	pattern := `^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`
	return regex.match(pattern, textStr)
}

func (Regex) isUUID(text any) bool {
	textStr := OSLtoString(text)
	pattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	return regex.match(pattern, textStr)
}

func (Regex) stripTags(text any) string {
	textStr := OSLtoString(text)
	pattern := `<[^>]*>`
	return regex.replace(pattern, textStr, "")
}

func (Regex) stripWhitespace(text any) string {
	textStr := OSLtoString(text)
	pattern := `\s+`
	return regex.replace(pattern, strings.TrimSpace(textStr), " ")
}

func (Regex) truncate(text any, length any, suffix any) string {
	textStr := OSLtoString(text)
	lengthVal := int(OSLcastNumber(length))
	suffixStr := OSLtoString(suffix)

	if len(textStr) <= lengthVal {
		return textStr
	}

	return textStr[:lengthVal] + suffixStr
}

func (Regex) slugify(text any) string {
	textStr := strings.ToLower(OSLtoString(text))

	slug := strings.ReplaceAll(textStr, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}

func (Regex) camelize(text any) string {
	textStr := OSLtoString(text)
	words := strings.Split(textStr, "_")

	for i, word := range words {
		if i > 0 {
			words[i] = strings.Title(word)
		}
	}

	return strings.Join(words, "")
}

func (Regex) snakeCase(text any) string {
	textStr := OSLtoString(text)

	match := regexp.MustCompile(`([a-z])([A-Z])`)
	result := match.ReplaceAllString(textStr, `${1}_${2}`)
	result = strings.ToLower(result)

	return result
}

func (Regex) kebabCase(text any) string {
	textStr := OSLtoString(text)

	match := regexp.MustCompile(`([a-z])([A-Z])`)
	result := match.ReplaceAllString(textStr, `${1}-${result}`)
	result = strings.ToLower(result)

	return strings.ReplaceAll(result, "_", "-")
}

func (Regex) maskEmail(email any) string {
	emailStr := OSLtoString(email)
	pattern := `(\S)(\S*@)`
	return regex.replace(pattern, emailStr, "*$2")
}

func (Regex) maskPhoneNumber(phone any) string {
	phoneStr := OSLtoString(phone)
	if len(phoneStr) < 4 {
		return phoneStr
	}

	masked := strings.Repeat("*", len(phoneStr)-4) + phoneStr[len(phoneStr)-4:]
	return masked
}

func (Regex) highlight(text any, pattern any, color any) string {
	textStr := OSLtoString(text)
	patternStr := OSLtoString(pattern)
	colorStr := OSLtoString(color)

	re := regexp.MustCompile(`(` + patternStr + `)`)
	result := re.ReplaceAllStringFunc(textStr, func(match string) string {
		return tui.Color(colorStr, match)
	})

	return result
}

func (Regex) wordCount(text any) int {
	textStr := OSLtoString(text)
	words := regex.extractWords(textStr)
	return len(words)
}

func (Regex) charCount(text any, includeSpaces any) int {
	textStr := OSLtoString(text)
	include := OSLcastBool(includeSpaces)

	if !include {
		textStr = strings.ReplaceAll(textStr, " ", "")
		return len(textStr)
	}

	return len(textStr)
}

func (Regex) sentenceCount(text any) int {
	textStr := OSLtoString(text)
	pattern := `[.!?]+`
	sentences := regex.split(pattern, textStr)

	count := 0
	for _, s := range sentences {
		if strings.TrimSpace(OSLtoString(s)) != "" {
			count++
		}
	}

	return count
}

var regex = Regex{}
