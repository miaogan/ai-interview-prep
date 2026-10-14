// Package graph contains Eino-based LLM chains.
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

// ErrInvalidResume marks "模型返回了 JSON，但内容不像一份简历"——映射为 40002。
// 其余错误（API 调用失败、输出不是合法 JSON）均为模型侧失败——映射为 50001。
var ErrInvalidResume = errors.New("简历内容无效")

const resumeSystemPrompt = `你是简历解析器。从用户给出的简历文本中抽取结构化画像，只输出一个 JSON 对象，不要输出任何解释或 markdown。
JSON 格式（字段名必须完全一致）：
{
  "name": "姓名",
  "yearsExp": 工作年限数字(应届填0),
  "targetRole": "目标岗位（依据求职意向或技能推断，如 AI应用工程师）",
  "skills": [{"name":"技能名","level":1到5的整数,"category":"language/framework/llm_app/finetune/mlops/algo/other 之一"}],
  "projects": [{"name":"项目名","techs":["技术1","技术2"],"desc":"一句话描述"}]
}
要求：
- level：5=精通 4=熟练 3=掌握 2=了解 1=接触过；依据简历中使用年限/项目深度判断，不要全给高分。
- skills 最多 12 个，按 level 从高到低排序；projects 最多 4 个。
- 只输出 JSON。`

// ParseResume runs the resume-parse chain: text → ChatModel(JSON) → validated Profile.
func ParseResume(ctx context.Context, cm model.ToolCallingChatModel, text string) (*domain.Profile, error) {
	out, err := cm.Generate(ctx, []*schema.Message{
		schema.SystemMessage(resumeSystemPrompt),
		schema.UserMessage("简历全文：\n" + text),
	})
	if err != nil {
		return nil, fmt.Errorf("模型调用失败: %w", err)
	}
	p, err := decodeProfile(out.Content)
	if err != nil {
		return nil, fmt.Errorf("模型输出无法解析为 JSON: %w", err)
	}
	if err := validateProfile(p); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResume, err)
	}
	return p, nil
}

// decodeProfile tolerates markdown fences / surrounding prose by slicing the
// first '{' to the last '}'.
func decodeProfile(content string) (*domain.Profile, error) {
	s := strings.TrimSpace(content)
	if i, j := strings.Index(s, "{"), strings.LastIndex(s, "}"); i >= 0 && j > i {
		s = s[i : j+1]
	}
	var p domain.Profile
	if err := json.Unmarshal([]byte(s), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func validateProfile(p *domain.Profile) error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("未识别到姓名")
	}
	skills := p.Skills[:0]
	for _, sk := range p.Skills {
		if strings.TrimSpace(sk.Name) == "" {
			continue
		}
		if sk.Level < 1 {
			sk.Level = 1
		}
		if sk.Level > 5 {
			sk.Level = 5
		}
		if sk.Category == "" {
			sk.Category = "other"
		}
		skills = append(skills, sk)
	}
	p.Skills = skills
	if len(p.Skills) == 0 && len(p.Projects) == 0 {
		return errors.New("未识别到任何技能或项目，可能不是一份简历")
	}
	if len(p.Skills) > 12 {
		p.Skills = p.Skills[:12]
	}
	if len(p.Projects) > 4 {
		p.Projects = p.Projects[:4]
	}
	if p.YearsExp < 0 {
		p.YearsExp = 0
	}
	if p.TargetRole == "" {
		p.TargetRole = "AI工程师"
	}
	return nil
}
