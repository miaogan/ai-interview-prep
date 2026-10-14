package api

import (
	"net/http"
	"sort"
	"strconv"
)

// 骨架阶段硬编码 feed（docs/03 步骤 5 将换为 MCP 检索）；relevanceDimension 用于按薄弱维度置顶。
var radarFeed = []RadarItem{
	{ID: "t1", Title: "vLLM 发布新调度优化，长上下文吞吐提升", Source: "GitHub", URL: "https://github.com/vllm-project/vllm", Summary: "改进的 KV cache 管理显著提升长文本推理吞吐。", PublishedAt: "2026-10-10", Tags: []string{"推理", "vLLM"}, RelevanceDimension: "mlops"},
	{ID: "t2", Title: "MCP 生态爆发：主流 Agent 框架全面支持", Source: "ModelContextProtocol", URL: "https://modelcontextprotocol.io/", Summary: "MCP 成为工具接入统一标准，Agent 平台标配。", PublishedAt: "2026-10-12", Tags: []string{"MCP", "Agent"}, RelevanceDimension: "llm_app"},
	{ID: "t3", Title: "开源模型 Qwen3 系列发布，长上下文能力增强", Source: "arXiv", URL: "https://arxiv.org/", Summary: "新一代开源模型在推理与长文本上进步明显。", PublishedAt: "2026-10-09", Tags: []string{"开源模型"}, RelevanceDimension: "dl_fundamentals"},
	{ID: "t4", Title: "DPO/偏好对齐新方法降低训练成本", Source: "arXiv", URL: "https://arxiv.org/", Summary: "无需显式奖励模型的对齐方法，适合资源受限微调。", PublishedAt: "2026-10-08", Tags: []string{"对齐", "微调"}, RelevanceDimension: "finetune"},
	{ID: "t5", Title: "GraphRAG 实践：图结构增强检索", Source: "GitHub", URL: "https://github.com/", Summary: "以图组织知识，提升多跳问答召回。", PublishedAt: "2026-10-11", Tags: []string{"RAG"}, RelevanceDimension: "llm_app"},
}

// radarItemsLocked 按当前薄弱维度置顶排序；调用方须持锁。
func (s *Server) radarItemsLocked() []RadarItem {
	weak := map[string]bool{}
	for _, rec := range s.computeGap().Recommendations {
		weak[rec.DimensionID] = true
	}
	items := append([]RadarItem(nil), radarFeed...)
	sort.SliceStable(items, func(i, j int) bool {
		return weak[items[i].RelevanceDimension] && !weak[items[j].RelevanceDimension]
	})
	return items
}

func (s *Server) handleRadarLatest(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	items := s.radarItemsLocked()
	s.mu.Unlock()

	if v := r.URL.Query().Get("limit"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit >= 0 && limit < len(items) {
			items = items[:limit]
		}
	}
	writeOK(w, RadarResp{Items: items, GeneratedAt: nowISO()})
}
