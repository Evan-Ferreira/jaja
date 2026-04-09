// Package skills provides data models and utilities for the ADK Skills system.
//
// Skills are folders of instructions and resources that extend agent
// capabilities for specialized tasks. Each skill directory contains a SKILL.md
// file with YAML frontmatter (L1), markdown instructions (L2), and optional
// subdirectories for references, assets, and scripts (L3).
package skills

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

var (
	kebabNamePattern    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	snakeOrKebabPattern = regexp.MustCompile(
		`^([a-z0-9]+(-[a-z0-9]+)*|[a-z0-9]+(_[a-z0-9]+)*)$`,
	)
)

// SkillDescriptor is implemented by types that can describe a skill
// (both Frontmatter and Skill satisfy this for use in FormatSkillsAsXML).
type SkillDescriptor interface {
	GetName() string
	GetDescription() string
}

// GetName implements SkillDescriptor.
func (f *Frontmatter) GetName() string { return f.Name }

// GetDescription implements SkillDescriptor.
func (f *Frontmatter) GetDescription() string { return f.Description }

// Frontmatter represents L1 skill content: metadata parsed from the YAML
// header of SKILL.md, used for lightweight skill discovery.
type Frontmatter struct {
	Name          string         `yaml:"name"            json:"name"`
	Description   string         `yaml:"description"     json:"description"`
	License       string         `yaml:"license,omitempty"       json:"license,omitempty"`
	Compatibility string         `yaml:"compatibility,omitempty" json:"compatibility,omitempty"`
	AllowedTools  string         `yaml:"allowed-tools,omitempty" json:"allowed-tools,omitempty"`
	Metadata      map[string]any `yaml:"metadata,omitempty"      json:"metadata,omitempty"`
}

// Validate checks that all frontmatter fields meet the agentskills.io spec.
// Set allowSnakeCase to true to also accept snake_case skill names.
func (f *Frontmatter) Validate(allowSnakeCase bool) error {
	// NOTE: full NFKC normalization requires golang.org/x/text/unicode/norm.
	// Omitted here to keep dependencies minimal; add norm.NFKC.String(f.Name)
	// if non-ASCII skill names are expected.

	if utf8.RuneCountInString(f.Name) > 64 {
		return fmt.Errorf("name must be at most 64 characters")
	}

	pattern := kebabNamePattern
	msg := ("name must be lowercase kebab-case (a-z, 0-9, hyphens), " +
		"with no leading, trailing, or consecutive delimiters")
	if allowSnakeCase {
		pattern = snakeOrKebabPattern
		msg = ("name must be lowercase kebab-case or snake_case, " +
			"with no leading, trailing, or consecutive delimiters; " +
			"mixing hyphens and underscores is not allowed")
	}
	
	if !pattern.MatchString(f.Name) {
		return fmt.Errorf("%s", msg)
	}

	if f.Description == "" {
		return fmt.Errorf("description must not be empty")
	}
	if utf8.RuneCountInString(f.Description) > 1024 {
		return fmt.Errorf("description must be at most 1024 characters")
	}

	if f.Compatibility != "" && utf8.RuneCountInString(f.Compatibility) > 500 {
		return fmt.Errorf("compatibility must be at most 500 characters")
	}

	if f.Metadata != nil {
		if tools, ok := f.Metadata["adk_additional_tools"]; ok {
			if _, isList := tools.([]any); !isList {
				return fmt.Errorf("adk_additional_tools must be a list of strings")
			}
		}
	}

	return nil
}

// Script wraps executable script content from a skill's scripts/ directory.
type Script struct {
	Src string `json:"src"`
}

func (s *Script) String() string { return s.Src }

// Resources represents L3 skill content: additional instructions, assets,
// and executable scripts loaded from subdirectories of a skill folder.
type Resources struct {
	References map[string][]byte  `json:"references,omitempty"`
	Assets     map[string][]byte  `json:"assets,omitempty"`
	Scripts    map[string]*Script `json:"scripts,omitempty"`
}

// NewResources returns an initialized Resources with empty maps.
func NewResources() *Resources {
	return &Resources{
		References: make(map[string][]byte),
		Assets:     make(map[string][]byte),
		Scripts:    make(map[string]*Script),
	}
}

// GetReference returns the content of a reference file, or nil if not found.
func (r *Resources) GetReference(id string) ([]byte, bool) {
	v, ok := r.References[id]
	return v, ok
}

// GetAsset returns the content of an asset file, or nil if not found.
func (r *Resources) GetAsset(id string) ([]byte, bool) {
	v, ok := r.Assets[id]
	return v, ok
}

// GetScript returns a Script object, or nil if not found.
func (r *Resources) GetScript(id string) (*Script, bool) {
	v, ok := r.Scripts[id]
	return v, ok
}

// ListReferences returns all reference file paths.
func (r *Resources) ListReferences() []string {
	keys := make([]string, 0, len(r.References))
	for k := range r.References {
		keys = append(keys, k)
	}
	return keys
}

// ListAssets returns all asset file paths.
func (r *Resources) ListAssets() []string {
	keys := make([]string, 0, len(r.Assets))
	for k := range r.Assets {
		keys = append(keys, k)
	}
	return keys
}

// ListScripts returns all script file paths.
func (r *Resources) ListScripts() []string {
	keys := make([]string, 0, len(r.Scripts))
	for k := range r.Scripts {
		keys = append(keys, k)
	}
	return keys
}

// Skill is the complete skill representation combining all three layers:
//   - L1: Frontmatter for discovery (name, description).
//   - L2: Instructions from SKILL.md body, loaded when the skill is triggered.
//   - L3: Resources (references, assets, scripts), loaded as needed.
type Skill struct {
	Frontmatter  *Frontmatter `json:"frontmatter"`
	Instructions string       `json:"instructions"`
	Resources    *Resources   `json:"resources,omitempty"`
}

// GetName implements SkillDescriptor.
func (s *Skill) GetName() string { return s.Frontmatter.Name }

// GetDescription implements SkillDescriptor.
func (s *Skill) GetDescription() string { return s.Frontmatter.Description }
