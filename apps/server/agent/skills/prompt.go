package skills

import (
	"html"
	"strings"
)

// FormatSkillsAsXML formats a list of skills into an XML string suitable
// for injection into LLM system instructions. Each skill's name and
// description are HTML-escaped.
func FormatSkillsAsXML(skills []SkillDescriptor) string {
	if len(skills) == 0 {
		return "<available_skills>\n</available_skills>"
	}

	var b strings.Builder
	b.WriteString("<available_skills>\n")

	for _, s := range skills {
		b.WriteString("<skill>\n")
		b.WriteString("<name>\n")
		b.WriteString(html.EscapeString(s.GetName()))
		b.WriteString("\n</name>\n")
		b.WriteString("<description>\n")
		b.WriteString(html.EscapeString(s.GetDescription()))
		b.WriteString("\n</description>\n")
		b.WriteString("</skill>\n")
	}

	b.WriteString("</available_skills>")
	return b.String()
}
