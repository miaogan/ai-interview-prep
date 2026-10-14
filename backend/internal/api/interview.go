package api

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// 脚本化多轮面试，对齐 docs/02；technical 第 3 题（索引 2）为「深入考察」，触发 HITL。
const deepDiveIndex = 2

var technicalScript = []string{
	"请用 2-3 分钟介绍一个你做过的 LLM 相关项目，以及你负责的模块。",
	"你提到用了 RAG，那检索召回不佳时你怎么排查和优化？",
	"我想深入考察向量库选型：HNSW 和 IVF 的区别与适用场景，以及 Rerank 放在哪一步？",
	"模型微调方面，LoRA 的 rank 设大了或设小了分别有什么问题？如何选？",
	"最后一题：如果要上线一个日均百万请求的推理服务，你会怎么设计保证低延迟？",
}

var behavioralScript = []string{
	"讲一次你和同事在技术方案上有分歧的经历，你怎么处理的？",
	"描述一个你在 deadline 压力下完成任务的例子。",
	"你最近主动学的新技术是什么？怎么学的？",
	"为什么想做 AI 开发工程师，你的规划是什么？",
}

// buildScript 按画像定制技术面 Q1/Q2：Q1 针对简历项目提问；项目技术栈不含 RAG 时 Q2 换通用深挖题。
func buildScript(typ string, p *Profile) []string {
	if typ != "technical" {
		return append([]string(nil), behavioralScript...)
	}
	s := append([]string(nil), technicalScript...)
	if p != nil && len(p.Projects) > 0 {
		proj := p.Projects[0]
		s[0] = fmt.Sprintf("你在《%s》中用了 %s，请介绍这个项目的整体架构、你负责的核心模块，以及最有挑战的部分。",
			proj.Name, strings.Join(proj.Techs, "、"))
		if !techsHit(proj.Techs, "rag", "langchain", "向量", "检索", "faiss") {
			s[1] = "在这个项目里，你遇到的最难的技术问题是什么？怎么定位和解决的？"
		}
	}
	return s
}

func techsHit(techs []string, keys ...string) bool {
	for _, t := range techs {
		for _, k := range keys {
			if strings.Contains(strings.ToLower(t), k) {
				return true
			}
		}
	}
	return false
}

// rubricPoint 一道题的单个考察要点：label 展示给用户，keys 用于命中检测。
type rubricPoint struct {
	label string
	keys  []string
}

var technicalRubric = [][]rubricPoint{
	{ // Q1 项目题
		rubricPoint{"技术栈与架构交代", []string{"架构", "技术栈", "框架", "langchain", "rag", "python", "pytorch", "faiss", "vllm", "检索", "向量"}},
		rubricPoint{"个人职责与模块", []string{"负责", "模块", "设计", "实现", "搭建", "开发", "主导"}},
		rubricPoint{"难点与挑战", []string{"挑战", "难点", "问题", "坑", "优化", "排查", "解决"}},
		rubricPoint{"量化成果", []string{"%", "提升", "降低", "qps", "延迟", "准确率", "覆盖率", "上线", "用户"}},
	},
	{ // Q2 RAG 排查
		rubricPoint{"Embedding 领域匹配", []string{"embedding", "向量模型", "领域", "bge", "m3e", "微调模型"}},
		rubricPoint{"Chunk 切分调优", []string{"chunk", "切分", "分块", "粒度"}},
		rubricPoint{"混合检索（关键词+向量）", []string{"混合检索", "bm25", "关键词", "hybrid", "稀疏"}},
		rubricPoint{"Rerank 精排", []string{"rerank", "重排", "精排"}},
		rubricPoint{"召回评估指标", []string{"召回率", "命中率", "hit", "评估", "指标", "badcase"}},
	},
	{ // Q3 HNSW/IVF 深入题
		rubricPoint{"HNSW 图索引特性", []string{"hnsw", "图索引", "图"}},
		rubricPoint{"IVF 倒排聚类", []string{"ivf", "聚类", "倒排"}},
		rubricPoint{"内存/延迟/召回权衡", []string{"内存", "延迟", "召回率", "权衡", "trade", "省"}},
		rubricPoint{"Rerank 所处阶段", []string{"rerank", "重排", "召回之后", "精排", "召回后"}},
	},
	{ // Q4 LoRA rank
		rubricPoint{"rank 过大的风险", []string{"过拟合", "显存", "遗忘", "参数多"}},
		rubricPoint{"rank 过小的局限", []string{"表达能力", "欠拟合", "容量", "学不到"}},
		rubricPoint{"取值经验与验证", []string{"8", "16", "32", "64", "验证集", "实验", "起步", "试"}},
	},
	{ // Q5 推理服务
		rubricPoint{"连续批处理", []string{"批处理", "batch", "continuous", "动态批"}},
		rubricPoint{"KV cache 管理", []string{"kv", "pagedattention", "显存"}},
		rubricPoint{"量化推理", []string{"量化", "int8", "int4", "fp8", "awq", "gptq"}},
		rubricPoint{"弹性扩容", []string{"扩容", "水平", "副本", "负载均衡", "弹性"}},
		rubricPoint{"缓存与 SLA 指标", []string{"缓存", "cache", "p99", "sla", "延迟要求", "吞吐"}},
	},
}

var behavioralRubric = [][]rubricPoint{
	{ // Q1 分歧处理
		rubricPoint{"倾听与理解对方", []string{"倾听", "理解", "对方", "立场", "沟通"}},
		rubricPoint{"用事实/数据论证", []string{"数据", "事实", "实验", "验证", "benchmark", "对比"}},
		rubricPoint{"达成共识的方式", []string{"共识", "妥协", "折中", "对齐", "讨论", "拍板"}},
		rubricPoint{"结果与反思", []string{"结果", "最终", "效果", "复盘", "教训"}},
	},
	{ // Q2 deadline 压力
		rubricPoint{"任务拆解与优先级", []string{"拆解", "优先级", "拆分", "里程碑", "排期"}},
		rubricPoint{"风险沟通与同步", []string{"沟通", "同步", "上报", "风险", "对齐"}},
		rubricPoint{"执行与按时交付", []string{"赶工", "加班", "交付", "按时", "完成", "上线"}},
		rubricPoint{"事后复盘", []string{"复盘", "改进", "总结", "教训"}},
	},
	{ // Q3 学习新技术
		rubricPoint{"学习动机", []string{"因为", "为了", "兴趣", "需要", "驱动", "工作"}},
		rubricPoint{"系统化方法", []string{"文档", "源码", "课程", "官方", "论文", "书"}},
		rubricPoint{"动手实践", []string{"实践", "项目", "demo", "练手", "落地", "动手"}},
		rubricPoint{"输出与沉淀", []string{"输出", "分享", "博客", "笔记", "总结", "沉淀"}},
	},
	{ // Q4 职业规划
		rubricPoint{"目标岗位认知", []string{"ai", "工程师", "岗位", "方向", "目标", "大模型"}},
		rubricPoint{"能力差距分析", []string{"差距", "短板", "不足", "提升", "弥补"}},
		rubricPoint{"具体行动计划", []string{"计划", "学习", "半年", "一年", "规划", "步骤"}},
		rubricPoint{"长期发展愿景", []string{"长期", "三年", "五年", "成长", "发展", "深耕"}},
	},
}

func rubricFor(typ string, idx int) []rubricPoint {
	r := behavioralRubric
	if typ == "technical" {
		r = technicalRubric
	}
	if idx < 0 || idx >= len(r) {
		return nil
	}
	return r[idx]
}

// annotate 对照考察要点批注回答；missed 最多 3 条，避免信息过载。
func annotate(answer string, points []rubricPoint) (covered, missed []string) {
	covered, missed = []string{}, []string{}
	a := strings.ToLower(answer)
	for _, p := range points {
		hit := false
		for _, k := range p.keys {
			if strings.Contains(a, strings.ToLower(k)) {
				hit = true
				break
			}
		}
		if hit {
			covered = append(covered, p.label)
		} else if len(missed) < 3 {
			missed = append(missed, p.label)
		}
	}
	return covered, missed
}

func makeNote(answer string, points []rubricPoint) (*AnswerNote, int) {
	if len(points) == 0 {
		return nil, 70
	}
	covered, missed := annotate(answer, points)
	tip := "要点覆盖全面，保持这个回答结构。"
	if len(missed) > 0 {
		tip = "可补充：" + strings.Join(missed, "、")
	}
	score := 55 + int(math.Round(45*float64(len(covered))/float64(len(points))))
	return &AnswerNote{Covered: covered, Missed: missed, Tip: tip}, score
}

type interviewState struct {
	sessionID  string
	typ        string
	script     []string
	transcript []Msg
	idx        int
	awaiting   bool
	finished   bool
	reviews    []QuestionReview
	scores     []int
	answers    []string
}

type createInterviewReq struct {
	Type       string `json:"type"`
	TargetRole string `json:"targetRole"`
	Rounds     int    `json:"rounds"`
}

func (s *Server) handleCreateInterview(w http.ResponseWriter, r *http.Request) {
	var req createInterviewReq
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Type != "technical" && req.Type != "behavioral" {
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "type must be technical | behavioral")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	script := buildScript(req.Type, s.profile)
	sessionID := fmt.Sprintf("s_%d", time.Now().UnixMilli())
	s.sessions[sessionID] = &interviewState{
		sessionID:  sessionID,
		typ:        req.Type,
		script:     script,
		transcript: []Msg{{Role: "interviewer", Content: script[0]}},
	}
	writeOK(w, InterviewSession{SessionID: sessionID, Type: req.Type, StartedAt: nowISO()})
}

type interviewMessageReq struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (s *Server) handleInterviewMessage(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.findSession(w, r)
	if !ok {
		return
	}
	var req interviewMessageReq
	if !decodeBody(w, r, &req) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if sess.finished {
		writeOK(w, InterviewMsgResp{Role: "interviewer", Content: "本场已结束。", Turn: sess.idx, Finished: true})
		return
	}
	// 前端用 __START__ 拉取首题：不入 transcript、不推进。
	if req.Content == "__START__" {
		writeOK(w, InterviewMsgResp{Role: "interviewer", Content: sess.transcript[0].Content, Turn: sess.idx, Finished: false})
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "content must not be empty")
		return
	}
	sess.transcript = append(sess.transcript, Msg{Role: "user", Content: req.Content})
	sess.answers = append(sess.answers, req.Content)

	// 对当前题的回答生成批注与覆盖度评分
	note, score := makeNote(req.Content, rubricFor(sess.typ, sess.idx))
	sess.reviews = append(sess.reviews, QuestionReview{
		Turn: sess.idx, Question: sess.script[sess.idx], Score: score,
		Covered: note.Covered, Missed: note.Missed,
	})
	sess.scores = append(sess.scores, score)
	fb := &Feedback{Score: score, Hint: note.Tip}

	// 答完第 2 题（idx=1）后，下一题是深入考察 → 先 HITL 暂停。
	if sess.typ == "technical" && sess.idx == 1 && !sess.awaiting {
		sess.awaiting = true
		writeOK(w, InterviewMsgResp{
			Role: "interviewer", Content: "", Turn: sess.idx, Finished: false,
			AwaitingApproval: true,
			ApprovalPrompt:   "面试官想就此深入考察技术细节，是否继续？",
			InterimFeedback:  fb, Note: note,
		})
		return
	}
	writeOK(w, s.advanceLocked(sess, fb, note))
}

func (s *Server) handleInterviewResume(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.findSession(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	sess.awaiting = false
	sess.idx = deepDiveIndex
	reply := sess.script[sess.idx]
	sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: reply})
	writeOK(w, InterviewMsgResp{
		Role: "interviewer", Content: reply, Turn: sess.idx, Finished: false,
		InterimFeedback: &Feedback{Score: 74, Hint: "继续展开。"},
	})
}

func (s *Server) handleInterviewSkip(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.findSession(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	sess.awaiting = false
	sess.idx = deepDiveIndex // 深入题记为已跳过，advance 到下一题
	writeOK(w, s.advanceLocked(sess, nil, nil))
}

func (s *Server) handleInterviewReport(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.findSession(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	overall := 0
	for _, sc := range sess.scores {
		overall += sc
	}
	if len(sess.scores) > 0 {
		overall = int(math.Round(float64(overall) / float64(len(sess.scores))))
	}

	report := InterviewReport{
		SessionID: sess.sessionID, Type: sess.typ, OverallScore: overall,
		Dimensions:   reportDimensions(sess),
		Highlights:   reviewSummary(sess.reviews, true),
		Improvements: reviewSummary(sess.reviews, false),
		Transcript:   sess.transcript,
		Review:       sess.reviews,
	}
	writeOK(w, report)
}

// advanceLocked 推进到下一题或结束面试；调用方须持锁。
func (s *Server) advanceLocked(sess *interviewState, fb *Feedback, note *AnswerNote) InterviewMsgResp {
	next := sess.idx + 1
	if next >= len(sess.script) {
		sess.finished = true
		sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: "好的，本轮面试结束，稍后给你评估报告。"})

		dim := "algo"
		if sess.typ == "technical" {
			dim = "llm_app"
		}
		s.memEvents = append(s.memEvents, memoryEvent{Type: "interview", DimensionID: dim, Delta: 3})

		overall := 0
		for _, sc := range sess.scores {
			overall += sc
		}
		if len(sess.scores) > 0 {
			overall = int(math.Round(float64(overall) / float64(len(sess.scores))))
		}
		s.intervHist = append(s.intervHist, interviewRecord{SessionID: sess.sessionID, Score: overall, At: nowISO()})
		return InterviewMsgResp{
			Role: "interviewer", Content: "感谢参与！面试已结束。", Turn: sess.idx, Finished: true,
			InterimFeedback: &Feedback{Score: overall, Hint: "报告已生成，去查看。"}, Note: note,
		}
	}

	sess.idx = next
	reply := sess.script[sess.idx]
	sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: reply})
	return InterviewMsgResp{Role: "interviewer", Content: reply, Turn: sess.idx, Finished: false, InterimFeedback: fb, Note: note}
}

// reportDimensions 三个评分维度由实际作答派生：整体均值、后半场（偏设计/规划）、表达结构。
func reportDimensions(sess *interviewState) []ScoreByName {
	avg := func(xs []int) int {
		if len(xs) == 0 {
			return 0
		}
		sum := 0
		for _, x := range xs {
			sum += x
		}
		return int(math.Round(float64(sum) / float64(len(xs))))
	}
	latter := sess.scores
	if len(latter) > 2 {
		latter = latter[len(latter)-2:]
	}

	// 表达/结构：平均作答长度 + STAR 结构词
	totalLen, starHits := 0, 0
	for _, a := range sess.answers {
		totalLen += len([]rune(a))
	}
	for _, w := range []string{"首先", "然后", "接着", "最后", "结果"} {
		for _, a := range sess.answers {
			if strings.Contains(a, w) {
				starHits++
				break
			}
		}
	}
	avgLen := 0
	if len(sess.answers) > 0 {
		avgLen = totalLen / len(sess.answers)
	}
	structure := 62
	switch {
	case avgLen >= 40:
		structure = 78
	case avgLen >= 20:
		structure = 70
	}
	if sess.typ == "behavioral" && starHits >= 2 {
		structure = min(85, structure+6)
	}

	if sess.typ == "technical" {
		return []ScoreByName{
			{Name: "技术深度", Score: avg(sess.scores)},
			{Name: "系统设计", Score: avg(latter)},
			{Name: "表达清晰度", Score: structure},
		}
	}
	return []ScoreByName{
		{Name: "沟通协作", Score: avg(sess.scores)},
		{Name: "职业规划", Score: avg(latter)},
		{Name: "STAR 结构", Score: structure},
	}
}

// reviewSummary 聚合逐题批注：亮点取覆盖多的轮次，改进取有缺失的轮次，各前 3 条。
func reviewSummary(reviews []QuestionReview, wantHighlight bool) []string {
	out := []string{}
	for _, rv := range reviews {
		if wantHighlight && len(rv.Covered) >= 2 {
			out = append(out, fmt.Sprintf("Q%d 覆盖：%s", rv.Turn+1, strings.Join(rv.Covered, "、")))
		}
		if !wantHighlight && len(rv.Missed) > 0 {
			out = append(out, fmt.Sprintf("Q%d 可补充：%s", rv.Turn+1, strings.Join(rv.Missed, "、")))
		}
		if len(out) == 3 {
			break
		}
	}
	if len(out) == 0 {
		if wantHighlight {
			out = []string{"全部题目均有作答，态度积极"}
		} else {
			out = []string{"各题要点覆盖全面，可挑战更高难度"}
		}
	}
	return out
}

func (s *Server) findSession(w http.ResponseWriter, r *http.Request) (*interviewState, bool) {
	id := r.PathValue("sessionId")
	s.mu.Lock()
	sess, ok := s.sessions[id]
	s.mu.Unlock()
	if !ok {
		writeErr(w, http.StatusNotFound, CodeNotFound, "interview session not found: "+id)
		return nil, false
	}
	return sess, true
}
