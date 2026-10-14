package api

import (
	"math"
	"net/http"
	"sort"
	"strings"
)

// 「AI 工程师能力模型」内置锚，见 docs/04；顺序即输出顺序。
type dimensionAnchor struct {
	ID     string
	Name   string
	Weight float64
	Skills []string
}

var dimensionAnchors = []dimensionAnchor{
	{ID: "dl_fundamentals", Name: "深度学习基础", Weight: 0.20, Skills: []string{"PyTorch", "TensorFlow", "CNN", "RNN", "Transformer"}},
	{ID: "llm_app", Name: "LLM应用(RAG/Agent)", Weight: 0.25, Skills: []string{"RAG", "LangChain", "Agent", "Prompt", "MCP", "向量检索"}},
	{ID: "finetune", Name: "模型微调", Weight: 0.20, Skills: []string{"LoRA", "QLoRA", "SFT", "RLHF", "DPO", "PEFT", "微调"}},
	{ID: "mlops", Name: "MLOps/部署", Weight: 0.15, Skills: []string{"Docker", "K8s", "量化", "vLLM", "部署", "服务化"}},
	{ID: "algo", Name: "算法与数据结构", Weight: 0.20, Skills: []string{"算法", "数据结构", "系统设计", "LeetCode"}},
}

var recommendationCatalog = map[string]Recommendation{
	"dl_fundamentals": {DimensionID: "dl_fundamentals", Topic: "Transformer 原理与手撕注意力",
		Resources: []Resource{{Title: "The Illustrated Transformer", URL: "https://jalammar.github.io/illustrated-transformer/", Type: "blog"}}},
	"llm_app": {DimensionID: "llm_app", Topic: "RAG 进阶：混合检索 + Rerank + Agent",
		Resources: []Resource{
			{Title: "LlamaIndex RAG", URL: "https://docs.llamaindex.ai/", Type: "doc"},
			{Title: "MCP 官方", URL: "https://modelcontextprotocol.io/", Type: "doc"},
		}},
	"finetune": {DimensionID: "finetune", Topic: "LoRA/QLoRA 微调实战",
		Resources: []Resource{{Title: "HuggingFace PEFT", URL: "https://huggingface.co/docs/peft", Type: "doc"}}},
	"mlops": {DimensionID: "mlops", Topic: "vLLM 部署 + 量化推理",
		Resources: []Resource{{Title: "vLLM 文档", URL: "https://docs.vllm.ai/", Type: "doc"}}},
	"algo": {DimensionID: "algo", Topic: "系统设计 + 手撕代码高频题",
		Resources: []Resource{{Title: "System Design Primer", URL: "https://github.com/donnemartin/system-design-primer", Type: "repo"}}},
}

// scoreDimension 技能匹配计分 + 记忆事件修正；调用方须持锁。
func (s *Server) scoreDimension(anchor dimensionAnchor) int {
	if s.profile == nil {
		return 0
	}
	sum := 0
	for _, skill := range s.profile.Skills {
		for _, a := range anchor.Skills {
			sk, al := strings.ToLower(skill.Name), strings.ToLower(a)
			if strings.Contains(sk, al) || strings.Contains(al, sk) {
				sum += skill.Level * 20
				break
			}
		}
	}
	mem := 0
	for _, e := range s.memEvents {
		if e.DimensionID == anchor.ID {
			mem += e.Delta
		}
	}
	return clampScore(sum + mem)
}

func clampScore(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func statusOf(score int) string {
	switch {
	case score >= 75:
		return "strong"
	case score >= 60:
		return "moderate"
	case score >= 40:
		return "weak"
	default:
		return "missing"
	}
}

// computeGap 调用方须持锁。
func (s *Server) computeGap() GapReport {
	dims := make([]Dimension, 0, len(dimensionAnchors))
	for _, a := range dimensionAnchors {
		score := s.scoreDimension(a)
		dims = append(dims, Dimension{ID: a.ID, Name: a.Name, Score: score, Weight: a.Weight, Status: statusOf(score)})
	}

	overall := 0.0
	for _, d := range dims {
		overall += float64(d.Score) * d.Weight
	}

	weak := make([]Dimension, 0, len(dims))
	for _, d := range dims {
		if d.Status == "weak" || d.Status == "missing" {
			weak = append(weak, d)
		}
	}
	sort.SliceStable(weak, func(i, j int) bool {
		return weak[i].Weight*float64(100-weak[i].Score) > weak[j].Weight*float64(100-weak[j].Score)
	})

	recs := make([]Recommendation, 0, len(weak))
	for _, d := range weak {
		rec := recommendationCatalog[d.ID]
		if d.Weight >= 0.2 && d.Status == "missing" {
			rec.Priority = "high"
		} else {
			rec.Priority = "medium"
		}
		recs = append(recs, rec)
	}

	return GapReport{
		OverallScore:    int(math.Round(overall)),
		Dimensions:      dims,
		Recommendations: recs,
	}
}

func (s *Server) handleGapAnalysis(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeOK(w, s.computeGap())
}
