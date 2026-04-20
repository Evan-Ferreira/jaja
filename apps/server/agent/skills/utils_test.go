package skills

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseSkillMDContent(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		content := "---\nname: my-skill\ndescription: A test skill\n---\n# Hello\nBody here."
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "name: my-skill\ndescription: A test skill" {
			t.Errorf("unexpected rawYAML: %q", rawYAML)
		}
		if body != "# Hello\nBody here." {
			t.Errorf("unexpected body: %q", body)
		}
	})

	t.Run("valid input with proper indentation in frontmatter", func(t *testing.T) {
		content := "---\nname: my-skill\ndescription: padded\n---\n  body text  "
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "name: my-skill\ndescription: padded" {
			t.Errorf("unexpected rawYAML: %q", rawYAML)
		}
		if body != "body text" {
			t.Errorf("unexpected body: %q", body)
		}
	})

	t.Run("empty body after frontmatter", func(t *testing.T) {
		content := "---\nname: my-skill\ndescription: A test skill\n---\n"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "name: my-skill\ndescription: A test skill" {
			t.Errorf("unexpected rawYAML: %q", rawYAML)
		}
		if body != "" {
			t.Errorf("expected empty body, got: %q", body)
		}
	})

	t.Run("empty body only whitespace", func(t *testing.T) {
		content := "---\nname: my-skill\ndescription: ok\n---\n   \n\t\n  "
		_, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if body != "" {
			t.Errorf("expected empty body after trimming, got: %q", body)
		}
	})
}

func TestParseSkillMDContent_MissingFrontmatter(t *testing.T) {
	t.Run("completely empty content", func(t *testing.T) {
		_, _, err := ParseSkillMDContent("")
		if err == nil {
			t.Fatal("expected error for empty content")
		}
		if !strings.Contains(err.Error(), "must start with YAML frontmatter") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("no frontmatter delimiters at all", func(t *testing.T) {
		_, _, err := ParseSkillMDContent("just some markdown text\nwith no frontmatter")
		if err == nil {
			t.Fatal("expected error for content without frontmatter")
		}
		if !strings.Contains(err.Error(), "must start with YAML frontmatter") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("frontmatter does not start at beginning", func(t *testing.T) {
		_, _, err := ParseSkillMDContent("\n---\nname: my-skill\n---\nbody")
		if err == nil {
			t.Fatal("expected error when --- is not at the very start")
		}
		if !strings.Contains(err.Error(), "must start with YAML frontmatter") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("space before opening delimiter", func(t *testing.T) {
		_, _, err := ParseSkillMDContent(" ---\nname: my-skill\n---\nbody")
		if err == nil {
			t.Fatal("expected error when --- has leading space")
		}
		if !strings.Contains(err.Error(), "must start with YAML frontmatter") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("opening delimiter only no closing", func(t *testing.T) {
		_, _, err := ParseSkillMDContent("---\nname: my-skill\ndescription: test\n")
		if err == nil {
			t.Fatal("expected error for unclosed frontmatter")
		}
		if !strings.Contains(err.Error(), "not properly closed") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestParseSkillMDContent_InvalidYAML(t *testing.T) {
	t.Run("malformed YAML in frontmatter", func(t *testing.T) {
		content := "---\n:\n  bad:\n    - [invalid\n---\nbody"
		_, _, err := ParseSkillMDContent(content)
		if err == nil {
			t.Fatal("expected error for malformed YAML")
		}
		if !strings.Contains(err.Error(), "invalid YAML") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("YAML with tabs instead of spaces", func(t *testing.T) {
		content := "---\nname: my-skill\n\tdescription: bad indent\n---\nbody"
		_, _, err := ParseSkillMDContent(content)
		if err == nil {
			t.Fatal("expected error for YAML with tabs")
		}
		if !strings.Contains(err.Error(), "invalid YAML") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("YAML that is a scalar not a mapping", func(t *testing.T) {
		// gopkg.in/yaml.v3 returns an error when unmarshaling a scalar
		// into map[string]any.
		content := "---\njust a plain string\n---\nbody"
		_, _, err := ParseSkillMDContent(content)
		if err == nil {
			t.Fatal("expected error for scalar YAML unmarshaled into map")
		}
		if !strings.Contains(err.Error(), "invalid YAML") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("YAML that is a list not a mapping", func(t *testing.T) {
		content := "---\n- item1\n- item2\n---\nbody"
		_, _, err := ParseSkillMDContent(content)
		if err == nil {
			t.Fatal("expected error for YAML list instead of mapping")
		}
		if !strings.Contains(err.Error(), "invalid YAML") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty frontmatter block", func(t *testing.T) {
		content := "---\n---\nbody text"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "" {
			t.Errorf("expected empty rawYAML, got: %q", rawYAML)
		}
		if body != "body text" {
			t.Errorf("unexpected body: %q", body)
		}
	})

	t.Run("frontmatter with only whitespace", func(t *testing.T) {
		content := "---\n   \n\n---\nbody"
		rawYAML, _, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "" {
			t.Errorf("expected empty rawYAML after trim, got: %q", rawYAML)
		}
	})

	t.Run("duplicate keys in YAML", func(t *testing.T) {
		// gopkg.in/yaml.v3 treats duplicate mapping keys as an error.
		content := "---\nname: first\nname: second\n---\nbody"
		_, _, err := ParseSkillMDContent(content)
		if err == nil {
			t.Fatal("expected error for duplicate YAML keys")
		}
		if !strings.Contains(err.Error(), "invalid YAML") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestParseSkillMDContent_EmptyContent(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		_, _, err := ParseSkillMDContent("")
		if err == nil {
			t.Fatal("expected error for empty string")
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		_, _, err := ParseSkillMDContent("   \n\t\n  ")
		if err == nil {
			t.Fatal("expected error for whitespace-only content")
		}
	})

	t.Run("only opening delimiter", func(t *testing.T) {
		_, _, err := ParseSkillMDContent("---")
		if err == nil {
			t.Fatal("expected error for only opening delimiter")
		}
		if !strings.Contains(err.Error(), "not properly closed") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("only delimiters no content between", func(t *testing.T) {
		content := "------"
		// "---" + "" + "---" — SplitN("------", "---", 3) -> ["", "", ""]
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "" {
			t.Errorf("expected empty rawYAML, got: %q", rawYAML)
		}
		if body != "" {
			t.Errorf("expected empty body, got: %q", body)
		}
	})
}

func TestParseSkillMDContent_LargeFiles(t *testing.T) {
	t.Run("very large body", func(t *testing.T) {
		largeBody := strings.Repeat("A", 1024*1024) // 1 MB of 'A's
		content := "---\nname: large-skill\ndescription: big\n---\n" + largeBody
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "name: large-skill\ndescription: big" {
			t.Errorf("unexpected rawYAML: %q", rawYAML)
		}
		if len(body) != 1024*1024 {
			t.Errorf("expected body length %d, got %d", 1024*1024, len(body))
		}
	})

	t.Run("large frontmatter with many keys", func(t *testing.T) {
		var sb strings.Builder
		sb.WriteString("---\n")
		for i := 0; i < 100; i++ {
			fmt.Fprintf(&sb, "key%d: value%d\n", i, i)
		}
		sb.WriteString("---\nbody")
		content := sb.String()
		_, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error for large frontmatter: %v", err)
		}
		if body != "body" {
			t.Errorf("unexpected body: %q", body)
		}
	})

	t.Run("moderately large file", func(t *testing.T) {
		largeBody := strings.Repeat("B", 256*1024) // 256 KB
		content := "---\nname: huge\ndescription: test\n---\n" + largeBody
		_, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(body) != 256*1024 {
			t.Errorf("expected body length %d, got %d", 256*1024, len(body))
		}
	})
}

func TestParseSkillMDContent_CharacterEncodings(t *testing.T) {
	t.Run("UTF-8 multibyte characters in body", func(t *testing.T) {
		content := "---\nname: i18n-skill\ndescription: Internationalized\n---\n# \xe4\xbd\xa0\xe5\xa5\xbd\xe4\xb8\x96\xe7\x95\x8c\n\nBody with emojis \xf0\x9f\x9a\x80\xf0\x9f\x8c\x8d and accents \xc3\xa9\xc3\xa0\xc3\xbc"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(rawYAML, "i18n-skill") {
			t.Errorf("rawYAML missing expected content: %q", rawYAML)
		}
		if !strings.Contains(body, "\xe4\xbd\xa0\xe5\xa5\xbd") {
			t.Errorf("body missing Chinese characters: %q", body)
		}
		if !strings.Contains(body, "\xf0\x9f\x9a\x80") {
			t.Errorf("body missing rocket emoji: %q", body)
		}
	})

	t.Run("UTF-8 multibyte characters in frontmatter values", func(t *testing.T) {
		content := "---\nname: multilang\ndescription: \xc3\x9cbersetzung f\xc3\xbcr Alle\n---\nbody"
		rawYAML, _, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(rawYAML, "\xc3\x9cbersetzung") {
			t.Errorf("rawYAML missing German characters: %q", rawYAML)
		}
	})

	t.Run("CJK characters throughout", func(t *testing.T) {
		content := "---\nname: cjk-test\ndescription: \xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e\xe3\x83\x86\xe3\x82\xb9\xe3\x83\x88\n---\n\xe3\x81\x93\xe3\x82\x8c\xe3\x81\xaf\xe3\x83\x86\xe3\x82\xb9\xe3\x83\x88\xe3\x81\xa7\xe3\x81\x99"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(rawYAML, "\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e") {
			t.Errorf("rawYAML missing Japanese: %q", rawYAML)
		}
		if !strings.Contains(body, "\xe3\x81\x93\xe3\x82\x8c\xe3\x81\xaf") {
			t.Errorf("body missing Japanese: %q", body)
		}
	})

	t.Run("content with BOM (byte order mark)", func(t *testing.T) {
		// UTF-8 BOM is 0xEF 0xBB 0xBF — prepended before ---
		content := "\xef\xbb\xbf---\nname: bom-skill\ndescription: has BOM\n---\nbody"
		_, _, err := ParseSkillMDContent(content)
		// BOM before --- means HasPrefix("---") fails
		if err == nil {
			t.Fatal("expected error when BOM precedes frontmatter delimiter")
		}
		if !strings.Contains(err.Error(), "must start with YAML frontmatter") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Latin-1 encoded bytes in body", func(t *testing.T) {
		// Simulates content that might not be valid UTF-8
		// \xff\xfe are invalid UTF-8 start bytes
		content := "---\nname: latin\ndescription: test\n---\nbody with \xff\xfe bytes"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "name: latin\ndescription: test" {
			t.Errorf("unexpected rawYAML: %q", rawYAML)
		}
		if !strings.Contains(body, "\xff\xfe") {
			t.Errorf("body should preserve raw bytes: %q", body)
		}
	})
}

func TestParseSkillMDContent_TripleDashEdgeCases(t *testing.T) {
	t.Run("triple dash in body content", func(t *testing.T) {
		content := "---\nname: my-skill\ndescription: test\n---\nSome text\n---\nMore text after separator"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rawYAML != "name: my-skill\ndescription: test" {
			t.Errorf("unexpected rawYAML: %q", rawYAML)
		}
		// SplitN with n=3 means only the first two --- are used as delimiters;
		// any additional --- in the body are preserved.
		if !strings.Contains(body, "---") {
			t.Errorf("body should contain --- separator: %q", body)
		}
	})

	t.Run("multiple triple dashes in body", func(t *testing.T) {
		content := "---\nname: test\ndescription: d\n---\npart1\n---\npart2\n---\npart3"
		_, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Count(body, "---") != 2 {
			t.Errorf("expected 2 occurrences of --- in body, got %d in: %q",
				strings.Count(body, "---"), body)
		}
	})

	t.Run("dashes longer than three", func(t *testing.T) {
		// "-----" contains "---" as prefix, SplitN splits on first "---"
		// leaving "--\nname: test\n..." as the YAML portion, which is
		// invalid YAML when unmarshaled into a map.
		content := "-----\nname: test\ndescription: d\n-----\nbody"
		_, _, err := ParseSkillMDContent(content)
		if err == nil {
			t.Fatal("expected error for malformed delimiter with extra dashes")
		}
		if !strings.Contains(err.Error(), "invalid YAML") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

func TestParseSkillMDContent_SpecialContent(t *testing.T) {
	t.Run("frontmatter with complex nested YAML", func(t *testing.T) {
		content := "---\nname: complex-skill\ndescription: nested\nmetadata:\n  adk_additional_tools:\n    - tool1\n    - tool2\n  custom:\n    nested_key: value\n---\nbody"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(rawYAML, "adk_additional_tools") {
			t.Errorf("rawYAML missing nested content: %q", rawYAML)
		}
		if body != "body" {
			t.Errorf("unexpected body: %q", body)
		}
	})

	t.Run("body with code blocks containing YAML", func(t *testing.T) {
		content := "---\nname: code-skill\ndescription: has code\n---\n# Instructions\n```yaml\nname: not-frontmatter\nvalue: 123\n```"
		_, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(body, "```yaml") {
			t.Errorf("body should contain code block: %q", body)
		}
	})

	t.Run("Windows-style line endings (CRLF)", func(t *testing.T) {
		content := "---\r\nname: crlf-skill\r\ndescription: windows\r\n---\r\nbody with CRLF"
		rawYAML, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(rawYAML, "crlf-skill") {
			t.Errorf("rawYAML should handle CRLF: %q", rawYAML)
		}
		if !strings.Contains(body, "body with CRLF") {
			t.Errorf("body should be present: %q", body)
		}
	})

	t.Run("null bytes in content", func(t *testing.T) {
		content := "---\nname: null-skill\ndescription: has nulls\n---\nbody with \x00 null byte"
		_, body, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(body, "\x00") {
			t.Errorf("body should preserve null bytes: %q", body)
		}
	})

	t.Run("unicode escape sequences in YAML", func(t *testing.T) {
		content := "---\nname: esc-skill\ndescription: \"Unicode \\u0041\\u0042\"\n---\nbody"
		rawYAML, _, err := ParseSkillMDContent(content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(rawYAML, "\\u0041") {
			t.Errorf("rawYAML should contain raw text: %q", rawYAML)
		}
	})
}
