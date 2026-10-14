package api

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

type examSession struct {
	exam Exam
}

// 题库与参考答案对齐 frontend mock.ts；判分规则：choice 比对，简答/编程关键词命中 >= min(2, len(keys))。
var questionBank = map[string][]Question{
	"llm_app": {
		{Type: "choice", DimensionID: "llm_app", Stem: "RAG 中用于缓解检索噪声、提升答案质量的关键策略是？",
			Options: []string{"A. 无限增大 top_k 直接拼接", "B. Rerank 对候选重排", "C. 提高采样温度", "D. 关闭向量检索"}},
		{Type: "coding", DimensionID: "llm_app", Stem: "用伪代码/Python 描述：你如何为向量库（千万级）设计低延迟检索方案？（关键词：ANN/IVF/HNSW、分片、Rerank、缓存）"},
	},
	"finetune": {
		{Type: "choice", DimensionID: "finetune", Stem: "QLoRA 相比标准 LoRA 的核心改进是？",
			Options: []string{"A. 全参数微调", "B. 4-bit 量化底座 + LoRA 适配器", "C. 去掉反向传播", "D. 只用 CPU 训练"}},
		{Type: "system_design", DimensionID: "finetune", Stem: "设计一个支持 DPO 偏好对齐的训练流水线，需覆盖数据、训练、评估、部署。（关键词：偏好数据、ref model、beta、奖励、评估集）"},
	},
	"dl_fundamentals": {
		{Type: "choice", DimensionID: "dl_fundamentals", Stem: "Transformer 中缩放点积注意力里除以 sqrt(dk) 的目的是？",
			Options: []string{"A. 加速收敛", "B. 防止点积过大导致 softmax 梯度过小", "C. 减少参数量", "D. 增加模型容量"}},
	},
	"mlops": {
		{Type: "choice", DimensionID: "mlops", Stem: "vLLM 提升推理吞吐的核心技术是？",
			Options: []string{"A. 静态 batching", "B. PagedAttention", "C. int8 训练", "D. 关闭 KV cache"}},
	},
	"algo": {
		{Type: "coding", DimensionID: "algo", Stem: "实现函数 topK(nums, k) 返回频率最高的 k 个元素。（关键词：哈希计数、堆/桶排序、O(n)）"},
	},
}

var refAnswers = map[string]struct {
	answer      string
	explanation string
}{
	"llm_app":         {"B", "Rerank 用交叉编码器对候选重排，过滤低相关片段，缓解噪声。"},
	"finetune":        {"B", "QLoRA 将底座量化到 4-bit 再训 LoRA 适配器，显存大幅下降。"},
	"dl_fundamentals": {"B", "除以 sqrt(dk) 防止点积过大让 softmax 进入饱和区、梯度消失。"},
	"mlops":           {"B", "PagedAttention 像操作系统分页一样管理 KV cache，显著提升吞吐。"},
}

var gradingKeywords = map[string][]string{
	"dl_fundamentals": {"softmax", "梯度", "dk", "方差", "缩放"},
	"llm_app":         {"rerank", "ann", "hnsw", "ivf", "分片", "缓存", "重排"},
	"finetune":        {"偏好", "ref", "beta", "dpo", "奖励", "评估"},
	"mlops":           {"paged", "attention", "kv", "吞吐", "batch"},
	"algo":            {"哈希", "堆", "桶", "o(n)", "count"},
}

func gradeByKeywords(dimensionID, answer string) bool {
	a := strings.ToLower(strings.TrimSpace(answer))
	if a == "" {
		return false
	}
	keys := gradingKeywords[dimensionID]
	hit := 0
	for _, k := range keys {
		if strings.Contains(a, strings.ToLower(k)) {
			hit++
		}
	}
	need := len(keys)
	if need > 2 {
		need = 2
	}
	return hit >= need
}

func (s *Server) handleGenerateExam(w http.ResponseWriter, r *http.Request) {
	var req ExamGenerateReq
	if !decodeBody(w, r, &req) {
		return
	}

	pool := make([]Question, 0)
	for _, dim := range req.DimensionIDs {
		for _, q := range questionBank[dim] {
			q.Difficulty = req.Difficulty
			pool = append(pool, q)
		}
	}
	count := req.Count
	if count <= 0 {
		count = 10
	}
	if count > len(pool) {
		count = len(pool)
	}

	questions := make([]Question, 0, count)
	for i := 0; i < count; i++ {
		q := pool[i]
		q.ID = fmt.Sprintf("q%d", i+1)
		questions = append(questions, q)
	}

	names := make([]string, 0, len(req.DimensionIDs))
	for _, dim := range req.DimensionIDs {
		for _, a := range dimensionAnchors {
			if a.ID == dim {
				names = append(names, strings.Split(a.Name, "(")[0])
				break
			}
		}
	}
	exam := Exam{
		ExamID:       fmt.Sprintf("e_%d", time.Now().UnixMilli()),
		Title:        strings.Join(names, " & ") + " · 专项笔试",
		TimeLimitSec: 1800,
		Questions:    questions,
	}

	s.mu.Lock()
	s.exam = &examSession{exam: exam}
	s.mu.Unlock()

	writeOK(w, exam)
}

type examSubmitReq struct {
	Answers map[string]string `json:"answers"`
}

func (s *Server) handleSubmitExam(w http.ResponseWriter, r *http.Request) {
	examID := r.PathValue("examId")
	var req examSubmitReq
	if !decodeBody(w, r, &req) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.exam == nil || s.exam.exam.ExamID != examID {
		writeErr(w, http.StatusNotFound, CodeNotFound, "exam not found: "+examID)
		return
	}

	results := make([]ResultItem, 0, len(s.exam.exam.Questions))
	correct := 0
	type dimStat struct{ correct, total int }
	stats := map[string]*dimStat{}
	statOrder := make([]string, 0)

	for _, q := range s.exam.exam.Questions {
		your := req.Answers[q.ID]
		var ref, expl string
		var isCorrect bool
		if q.Type == "choice" {
			ra, ok := refAnswers[q.DimensionID]
			if !ok {
				ra = struct{ answer, explanation string }{"A", ""}
			}
			ref, expl = ra.answer, ra.explanation
			isCorrect = strings.EqualFold(strings.TrimSpace(your), ref)
		} else {
			ref = "(关键词命中即得分)"
			expl = "参考要点：" + strings.Join(gradingKeywords[q.DimensionID], " / ")
			isCorrect = gradeByKeywords(q.DimensionID, your)
		}
		if isCorrect {
			correct++
		}
		results = append(results, ResultItem{
			ID: q.ID, YourAnswer: your, CorrectAnswer: ref,
			IsCorrect: isCorrect, Explanation: expl, DimensionID: q.DimensionID,
		})

		st, ok := stats[q.DimensionID]
		if !ok {
			st = &dimStat{}
			stats[q.DimensionID] = st
			statOrder = append(statOrder, q.DimensionID)
		}
		st.total++
		if isCorrect {
			st.correct++
		}

		delta := 2
		if !isCorrect {
			delta = -2
		}
		s.memEvents = append(s.memEvents, memoryEvent{Type: "exam", DimensionID: q.DimensionID, Delta: delta})
	}

	total := len(s.exam.exam.Questions)
	score := 0
	if total > 0 {
		score = int(math.Round(float64(correct) / float64(total) * 100))
	}

	weakPoints := make([]string, 0)
	for _, dim := range statOrder {
		st := stats[dim]
		if float64(st.correct)/float64(st.total) < 0.6 {
			weakPoints = append(weakPoints, dim)
		}
	}

	s.examHist = append(s.examHist, examRecord{ExamID: examID, Score: score, At: nowISO()})

	writeOK(w, ExamResult{
		ExamID: examID, Score: score, Total: 100,
		Results: results, WeakPoints: weakPoints, SubmittedAt: nowISO(),
	})
}
