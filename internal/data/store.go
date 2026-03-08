package data

import (
	"encoding/json"
	"fmt"
	"os"
)

// ── Profile ──
type Profile struct {
	Name         string        `json:"name"`
	Tagline      string        `json:"tagline"`
	Location     string        `json:"location"`
	Education    Education     `json:"education"`
	Contact      Contact       `json:"contact"`
	Stats        []Stat        `json:"stats"`
	Achievements []Achievement `json:"achievements"`
	Experience   []Experience  `json:"experience"`
	Projects     []Project     `json:"projects"`
	Skills       []SkillGroup  `json:"skills"`
}

type Education struct {
	Institution string `json:"institution"`
	Degree      string `json:"degree"`
	CGPA        string `json:"cgpa"`
	Period      string `json:"period"`
}

type Contact struct {
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	LinkedIn   string `json:"linkedin"`
	GitHub     string `json:"github"`
	CodeForces string `json:"codeforces"`
}

type Stat struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type Achievement struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ── Experience ──
type Experience struct {
	ID         string   `json:"id"`
	Role       string   `json:"role"`
	Company    string   `json:"company"`
	Location   string   `json:"location"`
	Period     string   `json:"period"`
	Highlights []string `json:"highlights"`
}

// ── Projects ──
type Project struct {
	Name          string            `json:"name"`
	RelatedSkills []string          `json:"related_skills"`
	Description   string            `json:"description"`
	Links         map[string]string `json:"links"`
}

// ── Skills ──
type SkillGroup struct {
	Category string   `json:"category"`
	Items    []string `json:"items"`
}

// Store holds the parsed JSON data in memory.
type Store struct {
	Profile Profile
}

// LoadAll reads all JSON files from the given root dir and populates the store.
func LoadAll(rootDir string) (*Store, error) {
	s := &Store{}

	if err := loadJSON(rootDir+"/data/profile.json", &s.Profile); err != nil {
		return nil, err
	}

	return s, nil
}

func loadJSON(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
