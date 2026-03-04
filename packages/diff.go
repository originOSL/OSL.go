// name: diff
// description: File and text diff utilities for comparing differences
// author: roturbot
// requires: strings

type Diff struct{}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Lines    []string
}

type DiffResult struct {
	Hunks   []DiffHunk
	OldText string
	NewText string
}

func NewDiff() Diff {
	return Diff{}
}

func (d Diff) compareLines(oldLines, newLines []string) DiffResult {
	result := DiffResult{}

	oldLen := len(oldLines)
	newLen := len(newLines)

	maxLines := oldLen
	if newLen > maxLines {
		maxLines = newLen
	}

	var currentHunk *DiffHunk
	oldCount := 0
	newCount := 0

	for i := 0; i < maxLines; i++ {
		oldLine := ""
		newLine := ""

		if i < oldLen {
			oldLine = oldLines[i]
		}
		if i < newLen {
			newLine = newLines[i]
		}

		if oldLine == newLine && currentHunk == nil {
			oldCount++
			newCount++
			continue
		}

		if currentHunk == nil {
			result.Hunks = append(result.Hunks, DiffHunk{
				OldStart: oldCount + 1,
				NewStart: newCount + 1,
			})
			currentHunk = &result.Hunks[len(result.Hunks)-1]
		}

		if i < oldLen && i < newLen {
			if oldLine != newLine {
				result.OldText += oldLine + "\n"
				result.NewText += newLine + "\n"
				currentHunk.Lines = append(currentHunk.Lines, "- "+oldLine)
				currentHunk.Lines = append(currentHunk.Lines, "+ "+newLine)
				currentHunk.OldLines++
				currentHunk.NewLines++
			} else {
				currentHunk.Lines = append(currentHunk.Lines, "  "+oldLine)
				currentHunk.OldLines++
				currentHunk.NewLines++
			}
		} else if i < oldLen {
			result.OldText += oldLine + "\n"
			currentHunk.Lines = append(currentHunk.Lines, "- "+oldLine)
			currentHunk.OldLines++
		} else if i < newLen {
			result.NewText += newLine + "\n"
			currentHunk.Lines = append(currentHunk.Lines, "+ "+newLine)
			currentHunk.NewLines++
		}

		oldCount++
		newCount++

		if i+1 < maxLines && (i+1 >= oldLen || i+1 >= newLen) {
			break
		}
	}

	return result
}

func (d Diff) compareWords(oldText, newText string) DiffResult {
	oldWords := d.tokenize(oldText)
	newWords := d.tokenize(newText)

	result := DiffResult{}
	var currentHunk *DiffHunk

	maxLen := len(oldWords)
	if len(newWords) > maxLen {
		maxLen = len(newWords)
	}

	for i := 0; i < maxLen; i++ {
		oldWord := ""
		newWord := ""

		if i < len(oldWords) {
			oldWord = oldWords[i]
		}
		if i < len(newWords) {
			newWord = newWords[i]
		}

		if oldWord == newWord && currentHunk == nil {
			continue
		}

		if currentHunk == nil {
			result.Hunks = append(result.Hunks, DiffHunk{
				OldStart: i + 1,
				NewStart: i + 1,
				OldLines: 0,
				NewLines: 0,
			})
			currentHunk = &result.Hunks[len(result.Hunks)-1]
		}

		if i < len(oldWords) && i < len(newWords) {
			if oldWord != newWord {
				result.OldText += oldWord + " "
				result.NewText += newWord + " "
				currentHunk.Lines = append(currentHunk.Lines, "-"+oldWord)
				currentHunk.Lines = append(currentHunk.Lines, "+"+newWord)
			} else {
				currentHunk.Lines = append(currentHunk.Lines, oldWord)
			}
		} else if i < len(oldWords) {
			result.OldText += oldWords[i] + " "
			currentHunk.Lines = append(currentHunk.Lines, "-"+oldWords[i])
		} else if i < len(newWords) {
			result.NewText += newWords[i] + " "
			currentHunk.Lines = append(currentHunk.Lines, "+"+newWords[i])
		}
	}

	return result
}

func (d Diff) tokenize(text string) []string {
	var words []string
	currentWord := ""

	for _, ch := range text {
		if ch == ' ' || ch == '\n' || ch == '\t' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(ch)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

func (d Diff) text(oldStr string, newStr string) DiffResult {
	newLines := strings.Split(newStr, "\n")
	oldLines := strings.Split(oldStr, "\n")

	return d.compareLines(oldLines, newLines)
}

func (d Diff) words(oldStr string, newStr string) DiffResult {
	return d.compareWords(oldStr, newStr)
}

func (d Diff) chars(oldStr string, newStr string) DiffResult {
	result := DiffResult{}

	oldRunes := []rune(oldStr)
	newRunes := []rune(newStr)

	maxLen := len(oldRunes)
	if len(newRunes) > maxLen {
		maxLen = len(newRunes)
	}

	var hunk *DiffHunk
	var oldText, newText strings.Builder

	for i := 0; i < maxLen; i++ {
		oldChar := ""
		newChar := ""

		if i < len(oldRunes) {
			oldChar = string(oldRunes[i])
		}
		if i < len(newRunes) {
			newChar = string(newRunes[i])
		}

		if oldChar == newChar && hunk == nil {
			continue
		}

		if hunk == nil {
			result.Hunks = append(result.Hunks, DiffHunk{
				OldStart: i + 1,
				NewStart: i + 1,
				OldLines: 0,
				NewLines: 0,
			})
			hunk = &result.Hunks[len(result.Hunks)-1]
		}

		if i < len(oldRunes) && i < len(newRunes) {
			if oldChar != newChar {
				oldText.WriteString(oldChar)
				newText.WriteString(newChar)
			}
		} else if i < len(oldRunes) {
			oldText.WriteString(oldChar)
		} else if i < len(newRunes) {
			newText.WriteString(newChar)
		}
	}

	result.OldText = oldText.String()
	result.NewText = newText.String()

	return result
}

func (d Diff) unified(result DiffResult) string {
	if len(result.Hunks) == 0 {
		return "No differences found"
	}

	var output strings.Builder

	for _, hunk := range result.Hunks {
		output.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
			hunk.OldStart, hunk.OldLines, hunk.NewStart, hunk.NewLines))

		for _, line := range hunk.Lines {
			output.WriteString(line + "\n")
		}
	}

	return output.String()
}

func (d Diff) html(result DiffResult) string {
	if len(result.Hunks) == 0 {
		return "<div class='diff'>No differences found</div>"
	}

	var output strings.Builder
	output.WriteString("<div class='diff'>\n")

	for _, hunk := range result.Hunks {
		for _, line := range hunk.Lines {
			if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "--") {
				output.WriteString(fmt.Sprintf("<div class='diff-old'>%s</div>\n", htmlEscapeString(line)))
			} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") {
				output.WriteString(fmt.Sprintf("<div class='diff-new'>%s</div>\n", htmlEscapeString(line)))
			} else {
				output.WriteString(fmt.Sprintf("<div class='diff-same'>%s</div>\n", htmlEscapeString(line)))
			}
		}
	}

	output.WriteString("</div>")
	return output.String()
}

func (d Diff) json(result DiffResult) string {
	jsonMap := map[string]any{
		"hunks":   result.Hunks,
		"oldText": result.OldText,
		"newText": result.NewText,
	}

	jsonString, _ := jsonMarshal(jsonMap)
	return OSLtoString(jsonString)
}

func htmlEscapeString(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

var diff = NewDiff()
