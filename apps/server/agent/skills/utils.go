package skills

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// ParseSkillMDContent splits raw SKILL.md content into the raw YAML
// frontmatter string and the markdown body. The content must begin with a
// YAML frontmatter block delimited by "---". The returned rawYAML is
// validated as a YAML mapping but not yet deserialized into a struct.
func ParseSkillMDContent(content string) (rawYAML string, body string, err error) {
	if !strings.HasPrefix(content, "---") {
		return "", "", errors.New(
			"SKILL.md must start with YAML frontmatter (---)",
		)
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return "", "", errors.New(
			"SKILL.md frontmatter not properly closed with ---",
		)
	}

	rawYAML = strings.TrimSpace(parts[1])

	var probe map[string]any
	if err := yaml.Unmarshal([]byte(rawYAML), &probe); err != nil {
		return "", "", fmt.Errorf("invalid YAML in frontmatter: %w", err)
	}

	body = strings.TrimSpace(parts[2])
	return rawYAML, body, nil
}

// parseFrontmatter unmarshals raw YAML directly into a validated Frontmatter.
func parseFrontmatter(rawYAML string) (*Frontmatter, error) {
	fm := &Frontmatter{Metadata: make(map[string]any)}

	if err := yaml.Unmarshal([]byte(rawYAML), fm); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	if err := fm.Validate(); err != nil {
		return nil, err
	}

	return fm, nil
}

// loadDir recursively reads all files from directory into a map keyed by
// relative path. Directories named __pycache__ are skipped. Files that are
// not valid UTF-8 are included as raw bytes.
func loadDir(dir string) (map[string][]byte, error) {
	files := make(map[string][]byte)

	info, err := os.Stat(dir)

	if err != nil || !info.IsDir() {
		return files, nil
	}

	err = filepath.WalkDir(dir, func(
		path string, d fs.DirEntry, walkErr error,
	) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		relPath, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		files[relPath] = data
		return nil
	})
	return files, err
}

// parseSkillMD resolves skillDir to an absolute path, then reads and parses
// its SKILL.md. Returns the resolved directory, raw YAML frontmatter, and body.
func parseSkillMD(
	skillDir string,
) (absDir string, rawYAML string, body string, err error) {
	absDir, err = filepath.Abs(skillDir)
	if err != nil {
		return "", "", "", err
	}

	info, statErr := os.Stat(absDir)
	if statErr != nil || !info.IsDir() {
		return "", "", "", fmt.Errorf(
			"skill directory '%s' not found", absDir,
		)
	}

	var mdPath string
	for _, name := range []string{"SKILL.md", "skill.md"} {
		candidate := filepath.Join(absDir, name)
		if _, err := os.Stat(candidate); err == nil {
			mdPath = candidate
			break
		}
	}

	if mdPath == "" {
		return "", "", "", fmt.Errorf(
			"SKILL.md not found in '%s'", absDir,
		)
	}

	data, readErr := os.ReadFile(mdPath)
	if readErr != nil {
		return "", "", "", fmt.Errorf("reading SKILL.md: %w", readErr)
	}

	rawYAML, body, parseErr := ParseSkillMDContent(string(data))
	if parseErr != nil {
		return "", "", "", parseErr
	}
	return absDir, rawYAML, body, nil
}

// LoadSkillFromDir loads a complete Skill from a local directory. The
// directory must contain a SKILL.md file whose frontmatter "name" field
// matches the directory name.
func LoadSkillFromDir(skillDir string) (*Skill, error) {
	absDir, rawYAML, body, err := parseSkillMD(skillDir)
	if err != nil {
		return nil, err
	}

	fm, err := parseFrontmatter(rawYAML)
	if err != nil {
		return nil, err
	}

	dirName := filepath.Base(absDir)
	if dirName != fm.Name {
		return nil, fmt.Errorf(
			"skill name '%s' does not match directory name '%s'",
			fm.Name, dirName,
		)
	}

	references, err := loadDir(filepath.Join(absDir, "references"))
	if err != nil {
		return nil, fmt.Errorf("loading references: %w", err)
	}
	assets, err := loadDir(filepath.Join(absDir, "assets"))
	if err != nil {
		return nil, fmt.Errorf("loading assets: %w", err)
	}
	rawScripts, err := loadDir(filepath.Join(absDir, "scripts"))
	if err != nil {
		return nil, fmt.Errorf("loading scripts: %w", err)
	}

	scripts := make(map[string]*Script, len(rawScripts))
	for name, content := range rawScripts {
		if !utf8.Valid(content) {
			continue
		}
		scripts[name] = &Script{Src: string(content)}
	}

	return &Skill{
		Frontmatter:  fm,
		Instructions: body,
		Resources: &Resources{
			References: references,
			Assets:     assets,
			Scripts:    scripts,
		},
	}, nil
}

// ReadSkillProperties reads only the frontmatter from a skill directory,
// without loading instructions or resources.
func ReadSkillProperties(skillDir string) (*Frontmatter, error) {
	_, rawYAML, _, err := parseSkillMD(skillDir)
	if err != nil {
		return nil, err
	}
	return parseFrontmatter(rawYAML)
}

// ListSkillsInDir lists all valid skills in a base directory. Each
// immediate subdirectory that contains a valid SKILL.md is included.
// Invalid skills are logged and skipped.
func ListSkillsInDir(basePath string) (map[string]*Frontmatter, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("skills base path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("skills base path '%s' is not a directory", absPath)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	result := make(map[string]*Frontmatter)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillFileName := entry.Name()
		fm, err := ReadSkillProperties(filepath.Join(absPath, skillFileName))
		if err != nil {
			continue
		}

		if _, ok := result[fm.Name]; ok {
			continue
		}
		
		result[fm.Name] = fm

	}

	return result, nil
}
