package api

import (
	"net/http"
	"sort"

	"aiprep/backend/internal/graph"
	llmmodel "aiprep/backend/internal/model"
)

// handleGeneratePlan 画像+Gap+记忆 → LLM 生成个性化学习路径（失败即报错，无静态兜底）。
func (s *Server) handleGeneratePlan(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.profile == nil {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "请先上传简历生成能力画像，再生成学习路径")
		return
	}
	gap := s.computeGap()
	in := graph.PlanInput{Profile: s.profile}
	for _, d := range gap.Dimensions {
		in.Dimensions = append(in.Dimensions, graph.PlanDimension{ID: d.ID, Name: d.Name, Score: d.Score, Status: d.Status})
	}
	sort.SliceStable(in.Dimensions, func(i, j int) bool { return in.Dimensions[i].Score < in.Dimensions[j].Score })
	// 最近记忆事件中的失分维度作为补充上下文
	seen := map[string]bool{}
	for i := len(s.memEvents) - 1; i >= 0 && len(in.RecentWeak) < 5; i-- {
		e := s.memEvents[i]
		if e.Delta < 0 && !seen[e.DimensionID] {
			seen[e.DimensionID] = true
			in.RecentWeak = append(in.RecentWeak, dimName(e.DimensionID)+"（"+e.Type+" 中失分）")
		}
	}
	s.mu.Unlock()

	cm, err := llmmodel.NewChatModel(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		return
	}
	plan, err := graph.BuildPlan(r.Context(), cm, in)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		return
	}
	plan.CreatedAt = nowISO()

	s.mu.Lock()
	s.plan = plan
	s.mu.Unlock()
	writeOK(w, plan)
}

func (s *Server) handleGetPlan(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeOK(w, s.plan) // 未生成时 data 为 null
}

func dimName(id string) string {
	for _, a := range dimensionAnchors {
		if a.ID == id {
			return a.Name
		}
	}
	return id
}
