package api

import "aiprep/backend/internal/domain"

// 契约结构体：与 docs/02、frontend/src/types/index.ts 一一对应（camelCase）。

// 画像类型定义在 domain 包（graph/resume 解析链共用），此处取别名保持 api 包既有用法不变。
type Skill = domain.Skill
type Project = domain.Project
type Profile = domain.Profile

type ResumeUploadResp struct {
	ResumeID      string   `json:"resumeId"`
	FileName      string   `json:"fileName"`
	ParsedProfile *Profile `json:"parsedProfile"`
	UploadedAt    string   `json:"uploadedAt"`
}

type Dimension struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Score  int     `json:"score"`
	Weight float64 `json:"weight"`
	Status string  `json:"status"` // strong | moderate | weak | missing
}

type Resource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type Recommendation struct {
	DimensionID string     `json:"dimensionId"`
	Priority    string     `json:"priority"` // high | medium | low
	Topic       string     `json:"topic"`
	Resources   []Resource `json:"resources"`
}

type GapReport struct {
	OverallScore    int              `json:"overallScore"`
	Dimensions      []Dimension      `json:"dimensions"`
	Recommendations []Recommendation `json:"recommendations"`
}

type Question struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"` // choice | coding | system_design
	DimensionID string   `json:"dimensionId"`
	Difficulty  string   `json:"difficulty"` // easy | medium | hard
	Stem        string   `json:"stem"`
	Options     []string `json:"options,omitempty"`
	Answer      *string  `json:"answer,omitempty"`
	Explanation *string  `json:"explanation,omitempty"`
}

type Exam struct {
	ExamID       string     `json:"examId"`
	Title        string     `json:"title"`
	TimeLimitSec int        `json:"timeLimitSec"`
	Questions    []Question `json:"questions"`
}

type ExamGenerateReq struct {
	DimensionIDs []string `json:"dimensionIds"`
	Count        int      `json:"count"`
	Types        []string `json:"types"`
	Difficulty   string   `json:"difficulty"`
}

type ResultItem struct {
	ID            string `json:"id"`
	YourAnswer    string `json:"yourAnswer"`
	CorrectAnswer string `json:"correctAnswer"`
	IsCorrect     bool   `json:"isCorrect"`
	Explanation   string `json:"explanation"`
	DimensionID   string `json:"dimensionId"`
}

type ExamResult struct {
	ExamID      string       `json:"examId"`
	Score       int          `json:"score"`
	Total       int          `json:"total"`
	Results     []ResultItem `json:"results"`
	WeakPoints  []string     `json:"weakPoints"`
	SubmittedAt string       `json:"submittedAt"`
}

type InterviewSession struct {
	SessionID string `json:"sessionId"`
	Type      string `json:"type"` // technical | behavioral
	StartedAt string `json:"startedAt"`
}

type Msg struct {
	Role    string `json:"role"` // interviewer | user
	Content string `json:"content"`
}

type Feedback struct {
	Score int    `json:"score"`
	Hint  string `json:"hint"`
}

// AnswerNote 每轮回答的批注：已覆盖要点、可补充要点、调整建议。
type AnswerNote struct {
	Covered []string `json:"covered"`
	Missed  []string `json:"missed"`
	Tip     string   `json:"tip"`
}

// QuestionReview 逐题复盘：题目、得分、覆盖/缺失要点。
type QuestionReview struct {
	Turn     int      `json:"turn"`
	Question string   `json:"question"`
	Score    int      `json:"score"`
	Covered  []string `json:"covered"`
	Missed   []string `json:"missed"`
}

type InterviewMsgResp struct {
	Role             string      `json:"role"` // 恒为 interviewer
	Content          string      `json:"content"`
	Turn             int         `json:"turn"`
	Finished         bool        `json:"finished"`
	AwaitingApproval bool        `json:"awaitingApproval,omitempty"`
	ApprovalPrompt   string      `json:"approvalPrompt,omitempty"`
	InterimFeedback  *Feedback   `json:"interimFeedback,omitempty"`
	Note             *AnswerNote `json:"note,omitempty"`
}

type InterviewReport struct {
	SessionID    string           `json:"sessionId"`
	Type         string           `json:"type"`
	OverallScore int              `json:"overallScore"`
	Dimensions   []ScoreByName    `json:"dimensions"`
	Highlights   []string         `json:"highlights"`
	Improvements []string         `json:"improvements"`
	Transcript   []Msg            `json:"transcript"`
	Review       []QuestionReview `json:"review"`
}

type ScoreByName struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type RadarItem struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Source             string   `json:"source"`
	URL                string   `json:"url"`
	Summary            string   `json:"summary"`
	PublishedAt        string   `json:"publishedAt"`
	Tags               []string `json:"tags"`
	RelevanceDimension string   `json:"relevanceDimension"`
}

type RadarResp struct {
	Items       []RadarItem `json:"items"`
	GeneratedAt string      `json:"generatedAt"`
}

type DashboardSummary struct {
	OverallScore     int               `json:"overallScore"`
	WeakDimensions   []string          `json:"weakDimensions"`
	RecentExams      []examRecord      `json:"recentExams"`
	RecentInterviews []interviewRecord `json:"recentInterviews"`
	StudyStreakDays  int               `json:"studyStreakDays"`
	RadarHighlights  []TitleURL        `json:"radarHighlights"`
}

type TitleURL struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}
