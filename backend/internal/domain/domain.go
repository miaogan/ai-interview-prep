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

type Resource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

// PlanItem 一个可执行的学习任务；Acceptance 是验收标准（通常联动笔试/面试）。
type PlanItem struct {
	DimensionID string     `json:"dimensionId"`
	Topic       string     `json:"topic"`
	Goal        string     `json:"goal"`
	Acceptance  string     `json:"acceptance"`
	EstDays     int        `json:"estDays"`
	Resources   []Resource `json:"resources"`
}

type PlanPhase struct {
	Name  string     `json:"name"`
	Weeks string     `json:"weeks"`
	Items []PlanItem `json:"items"`
}

type StudyPlan struct {
	CreatedAt  string      `json:"createdAt"`
	TargetRole string      `json:"targetRole"`
	TotalWeeks int         `json:"totalWeeks"`
	Phases     []PlanPhase `json:"phases"`
}
