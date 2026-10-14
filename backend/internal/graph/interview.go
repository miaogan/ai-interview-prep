package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"aiprep/backend/internal/domain"
)

// InterviewTurnResult 一轮评估的输出：批注 + 评分 + 下一题（或深入考察提议）。
type InterviewTurnResult struct {
	Score            int      `json:"score"`
	Covered          []string `json:"covered"`
	Missed           []string `json:"missed"`
	Tip              string   `json:"tip"`
	NextQuestion     string   `json:"nextQuestion"`
	DeepDiveProposal string   `json:"deepDiveProposal"`
	Finished         bool     `json:"finished"`
}

type InterviewReportResult struct {
	OverallScore int                  `json:"overallScore"`
	Dimensions   []domain.ScoreByName `json:"dimensions"`
	Highlights   []string             `json:"highlights"`
	Improvements []string             `json:"improvements"`
}

// QA 一对已完成的问答。
type QA struct {
	Question string
	Answer   string
}

var ErrInvalidInterview = errors.New("面试模型输出无效")

func intervieweeProfile(p *domain.Profile) string {
	if p == nil {
		return "（候选人未提供简历，按通用背景提问）"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "姓名 %s，%d 年经验，目标岗位 %s。\n技能：", p.Name, p.YearsExp, p.TargetRole)
	for i, sk := range p.Skills {
		if i >= 10 {
			break
		}
		fmt.Fprintf(&sb, "%s(L%d) ", sk.Name, sk.Level)
	}
	sb.WriteString("\n项目：")
	for i, pr := range p.Projects {
		if i >= 3 {
			break
		}
		fmt.Fprintf(&sb, "《%s》(%s) %s；", pr.Name, strings.Join(pr.Techs, "/"), pr.Desc)
	}
	return sb.String()
}

func transcriptText(qas []QA) string {
	var sb strings.Builder
	for i, qa := range qas {
		fmt.Fprintf(&sb, "Q%d：%s\nA%d：%s\n", i+1, qa.Question, i+1, qa.Answer)
	}
	return sb.String()
}

func typeName(typ string) string {
	if typ == "technical" {
		return "技术"
	}
	return "HR 行为"
}

func generateJSON(ctx context.Context, cm model.ToolCallingChatModel, sys, user string, dst any) error {
	out, err := cm.Generate(ctx, []*schema.Message{
		schema.SystemMessage(sys),
		schema.UserMessage(user),
	})
	if err != nil {
		return fmt.Errorf("模型调用失败: %w", err)
	}
	s := strings.TrimSpace(out.Content)
	if i, j := strings.Index(s, "{"), strings.LastIndex(s, "}"); i >= 0 && j > i {
		s = s[i : j+1]
	}
	if err := json.Unmarshal([]byte(s), dst); err != nil {
		return fmt.Errorf("模型输出无法解析为 JSON: %w", err)
	}
	return nil
}

// OpeningQuestion 生成首题：技术面针对简历项目，行为面围绕协作/抗压。
func OpeningQuestion(ctx context.Context, cm model.ToolCallingChatModel, p *domain.Profile, typ string) (string, error) {
	sys := fmt.Sprintf(`你是资深%s面试官，正在面试一位应聘者的第一题。
候选人画像：
%s
要求：技术面首题必须针对其某个具体项目（整体架构、个人职责、最有挑战的部分）；行为面围绕团队协作或抗压经历。只问一个问题，具体、可回答。
只输出 JSON：{"question": "..."}`, typeName(typ), intervieweeProfile(p))
	var r struct {
		Question string `json:"question"`
	}
	if err := generateJSON(ctx, cm, sys, "请出题。", &r); err != nil {
		return "", err
	}
	if strings.TrimSpace(r.Question) == "" {
		return "", fmt.Errorf("%w: 开场题为空", ErrInvalidInterview)
	}
	return strings.TrimSpace(r.Question), nil
}

// EvaluateTurn 评价当前回答并生成下一题。turnNo/totalTurns 从 1 计。
func EvaluateTurn(ctx context.Context, cm model.ToolCallingChatModel, p *domain.Profile, typ string, qas []QA, question, answer string, turnNo, totalTurns int, deepDiveUsed bool) (*InterviewTurnResult, error) {
	deepRule := "本场尚未深入考察：若回答中有值得深挖的技术点，给出 deepDiveProposal（一句话说明想深挖什么），否则为 null。"
	if deepDiveUsed || typ != "technical" {
		deepRule = "deepDiveProposal 一律为 null。"
	}
	lastRule := "评价完后 finished=false，并给出 nextQuestion。"
	if turnNo >= totalTurns {
		lastRule = "这是最后一题，评价完后 finished=true、nextQuestion 为 null。"
	}
	sys := fmt.Sprintf(`你是资深%s面试官，正在面试应聘者（第 %d/%d 题）。
候选人画像：
%s
已完成问答：
%s
当前问题：%s
候选人回答：%s

任务：
1. 评价回答：score 0-100（要点覆盖、深度、准确性综合；拉开区分度，不要集中在 70-85）
2. 批注：covered=回答覆盖的要点（1-4 条），missed=遗漏的关键点（最多 3 条），tip=一句话可执行的改进建议
3. 出题：nextQuestion 基于回答内容追问或切换到新主题；一次只问一个具体问题，避免空泛
%s
%s
只输出 JSON：{"score":0,"covered":[],"missed":[],"tip":"","nextQuestion":null,"deepDiveProposal":null,"finished":false}`,
		typeName(typ), turnNo, totalTurns, intervieweeProfile(p), transcriptText(qas), question, answer, deepRule, lastRule)
	var r InterviewTurnResult
	if err := generateJSON(ctx, cm, sys, "请评价并出题。", &r); err != nil {
		return nil, err
	}
	if r.Score < 0 {
		r.Score = 0
	}
	if r.Score > 100 {
		r.Score = 100
	}
	if len(r.Missed) > 3 {
		r.Missed = r.Missed[:3]
	}
	if !r.Finished && strings.TrimSpace(r.NextQuestion) == "" {
		return nil, fmt.Errorf("%w: 未结束但 nextQuestion 为空", ErrInvalidInterview)
	}
	if turnNo >= totalTurns {
		r.Finished = true
	}
	return &r, nil
}

// DeepDiveQuestion 用户同意深入考察后，针对提议点生成深挖题。
func DeepDiveQuestion(ctx context.Context, cm model.ToolCallingChatModel, p *domain.Profile, typ string, qas []QA, point string) (string, error) {
	sys := fmt.Sprintf(`你是资深%s面试官。基于面试记录，围绕「%s」提出一个更深入的追问，考察理解深度（原理、权衡、取舍），而不是重复已问内容。
候选人画像：
%s
已完成问答：
%s
只输出 JSON：{"question": "..."}`, typeName(typ), point, intervieweeProfile(p), transcriptText(qas))
	var r struct {
		Question string `json:"question"`
	}
	if err := generateJSON(ctx, cm, sys, "请出深入考察题。", &r); err != nil {
		return "", err
	}
	if strings.TrimSpace(r.Question) == "" {
		return "", fmt.Errorf("%w: 深挖题为空", ErrInvalidInterview)
	}
	return strings.TrimSpace(r.Question), nil
}

// EvaluateInterview 读完整 transcript 产出总评；review 逐题数据由后端已有批注组装，不走模型。
func EvaluateInterview(ctx context.Context, cm model.ToolCallingChatModel, p *domain.Profile, typ string, qas []QA) (*InterviewReportResult, error) {
	dims := "技术深度 / 系统设计 / 表达清晰度"
	if typ != "technical" {
		dims = "沟通协作 / 职业规划 / STAR 结构"
	}
	sys := fmt.Sprintf(`你是面试评估专家。根据完整面试记录输出评估报告。
候选人画像：
%s
面试记录：
%s
要求：
- overallScore 0-100，与各题表现一致，拉开区分度
- dimensions 固定为：%s，各 0-100
- highlights / improvements 各 2-4 条，必须引用面试中的具体内容，禁止空话（如"态度积极"）
只输出 JSON：{"overallScore":0,"dimensions":[{"name":"","score":0}],"highlights":[],"improvements":[]}`,
		intervieweeProfile(p), transcriptText(qas), dims)
	var r InterviewReportResult
	if err := generateJSON(ctx, cm, sys, "请输出评估报告。", &r); err != nil {
		return nil, err
	}
	if r.OverallScore < 0 || r.OverallScore > 100 {
		return nil, fmt.Errorf("%w: overallScore 越界 %d", ErrInvalidInterview, r.OverallScore)
	}
	if len(r.Dimensions) == 0 {
		return nil, fmt.Errorf("%w: dimensions 为空", ErrInvalidInterview)
	}
	for i := range r.Dimensions {
		if r.Dimensions[i].Score < 0 {
			r.Dimensions[i].Score = 0
		}
		if r.Dimensions[i].Score > 100 {
			r.Dimensions[i].Score = 100
		}
	}
	if len(r.Highlights) == 0 || len(r.Improvements) == 0 {
		return nil, fmt.Errorf("%w: highlights/improvements 为空", ErrInvalidInterview)
	}
	return &r, nil
}
