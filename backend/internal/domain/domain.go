// Package domain holds core data structures shared by api, graph and resume packages.
package domain

type Skill struct {
	Name     string `json:"name"`
	Level    int    `json:"level"` // 1-5
	Category string `json:"category"`
}

type Project struct {
	Name  string   `json:"name"`
	Techs []string `json:"techs"`
	Desc  string   `json:"desc"`
}

type Profile struct {
	Name       string    `json:"name"`
	YearsExp   int       `json:"yearsExp"`
	TargetRole string    `json:"targetRole"`
	Skills     []Skill   `json:"skills"`
	Projects   []Project `json:"projects"`
	UpdatedAt  string    `json:"updatedAt,omitempty"`
}
