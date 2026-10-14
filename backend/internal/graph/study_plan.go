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

// PlanInput 学习路径生成的全部上下文。
type PlanInput struct {
	Profile    *domain.Profile
	Dimensions []PlanDimension // 已按薄弱程度排序
	RecentWeak []string        // 最近练习/面试暴露的薄弱点描述
}

type PlanDimension struct {
	ID     string
	Name   string
	Score  int
	Status string
}

// ErrInvalidPlan 表示模型返回了 JSON 但结构/内容不可用——模型侧失败，映射 50001。
var ErrInvalidPlan = errors.New("学习路径内容无效")

const planSystemPrompt = `你是求职备考教练。根据用户的能力画像和各维度得分，生成一份个性化学习路径，只输出一个 JSON 对象，不要输出任何解释或 markdown。
JSON 格式（字段名必须完全一致）：
{
  "totalWeeks": 总周数(4到8的整数),
  "phases": [
    {
      "name": "阶段名（如 阶段一：补齐算法短板）",
      "weeks": "第1-2周",
      "items": [
        {
          "dimensionId": "dl_fundamentals/llm_app/finetune/mlops/algo 之一",
          "topic": "学习主题",
          "goal": "具体可衡量的目标",
          "acceptance": "验收标准（可量化，例如：笔试该维度达到70分；能完整讲清某两项技术的取舍）",
          "estDays": 预计天数(1到14的整数),
          "resources": [{"title":"资源名","url":"https://...","type":"doc/blog/repo/video"}]
        }
      ]
    }
  ]
}
要求：
- 优先安排得分最低、权重最高的维度；每个阶段 1~3 个 item，总共 2~4 个阶段。
- 已扎实的维度（strong）不要安排。
- resources 只用知名官方文档/经典仓库/公认教程（如官方文档、GitHub 知名仓库），不要编造冷门链接；每个 item 1~2 条。
- 目标要贴合用户的现有技能与目标岗位，避免泛泛而谈。
- 只输出 JSON。`

// BuildPlan runs: context → ChatModel(JSON) → cleaned StudyPlan.
func BuildPlan(ctx context.Context, cm model.ToolCallingChatModel, in PlanInput) (*domain.StudyPlan, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "目标岗位：%s，工作年限：%d 年\n", in.Profile.TargetRole, in.Profile.YearsExp)
	sb.WriteString("已掌握技能：")
	for i, sk := range in.Profile.Skills {
		if i >= 12 {
			break
		}
		fmt.Fprintf(&sb, "%s(L%d) ", sk.Name, sk.Level)
	}
	sb.WriteString("\n维度得分（100 满分，按薄弱程度排序）：\n")
	for _, d := range in.Dimensions {
		fmt.Fprintf(&sb, "- %s(%s): %d 分 [%s]\n", d.Name, d.ID, d.Score, d.Status)
	}
	if len(in.RecentWeak) > 0 {
		sb.WriteString("最近练习/面试暴露的薄弱点：\n")
		for _, w := range in.RecentWeak {
			sb.WriteString("- " + w + "\n")
		}
	}

	out, err := cm.Generate(ctx, []*schema.Message{
		schema.SystemMessage(planSystemPrompt),
		schema.UserMessage(sb.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("模型调用失败: %w", err)
	}
	plan, err := decodePlan(out.Content)
	if err != nil {
		return nil, fmt.Errorf("模型输出无法解析为 JSON: %w", err)
	}
	if err := validatePlan(plan); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPlan, err)
	}
	plan.TargetRole = in.Profile.TargetRole
	return plan, nil
}

func decodePlan(content string) (*domain.StudyPlan, error) {
	s := strings.TrimSpace(content)
	if i, j := strings.Index(s, "{"), strings.LastIndex(s, "}"); i >= 0 && j > i {
		s = s[i : j+1]
	}
	var p domain.StudyPlan
	if err := json.Unmarshal([]byte(s), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

var validPlanDims = map[string]bool{
	"dl_fundamentals": true, "llm_app": true, "finetune": true, "mlops": true, "algo": true,
}

func validatePlan(p *domain.StudyPlan) error {
	phases := p.Phases[:0]
	for _, ph := range p.Phases {
		items := ph.Items[:0]
		for _, it := range ph.Items {
			if strings.TrimSpace(it.Topic) == "" || !validPlanDims[it.DimensionID] {
				continue
			}
			if it.EstDays < 1 {
				it.EstDays = 1
			}
			if it.EstDays > 14 {
				it.EstDays = 14
			}
			items = append(items, it)
		}
		ph.Items = items
		if len(ph.Items) > 0 {
			phases = append(phases, ph)
		}
	}
	p.Phases = phases
	if len(p.Phases) == 0 {
		return errors.New("没有任何有效学习项")
	}
	if p.TotalWeeks < 1 {
		p.TotalWeeks = 4
	}
	if p.TotalWeeks > 12 {
		p.TotalWeeks = 12
	}
	return nil
}
