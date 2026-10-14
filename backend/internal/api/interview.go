package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"aiprep/backend/internal/graph"
	llmmodel "aiprep/backend/internal/model"

	"github.com/cloudwego/eino/components/model"
)

// LLM 驱动的多轮面试：每轮一次模型调用，产出语义批注 + 动态追问；HITL 由模型提议深入点。
type interviewState struct {
	sessionID  string
	typ        string
	totalTurns int

	transcript []Msg
	qas        []graph.QA
	currentQ   string

	awaiting        bool
	finished        bool
	busy            bool // 有 LLM 调用在途，防重入
	deepDiveUsed    bool
	pendingNext     string
	pendingDeepDive string

	reviews []QuestionReview
	scores  []int
	report  *InterviewReport
}

func totalTurnsOf(typ string) int {
	if typ == "technical" {
		return 5
	}
	return 4
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
	sessionID := fmt.Sprintf("s_%d", time.Now().UnixMilli())
	s.sessions[sessionID] = &interviewState{sessionID: sessionID, typ: req.Type, totalTurns: totalTurnsOf(req.Type)}
	writeOK(w, InterviewSession{SessionID: sessionID, Type: req.Type, StartedAt: nowISO()})
}

type interviewMessageReq struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatModelOrErr 构建 ChatModel，失败时写 50001。
func chatModelOrErr(w http.ResponseWriter, r *http.Request) (model.ToolCallingChatModel, bool) {
	cm, err := llmmodel.NewChatModel(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		return nil, false
	}
	return cm, true
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
	if sess.finished {
		s.mu.Unlock()
		writeOK(w, InterviewMsgResp{Role: "interviewer", Content: "本场已结束。", Turn: len(sess.scores), Finished: true})
		return
	}

	// 前端用 __START__ 拉取首题：已拉取过则幂等返回，否则 LLM 生成开场题。
	if req.Content == "__START__" {
		if sess.currentQ != "" {
			s.mu.Unlock()
			writeOK(w, InterviewMsgResp{Role: "interviewer", Content: sess.currentQ, Turn: 0, Finished: false})
			return
		}
		if sess.busy {
			s.mu.Unlock()
			writeErr(w, http.StatusBadRequest, CodeBadRequest, "开场题生成中，请稍候")
			return
		}
		sess.busy = true
		profile, typ := s.profile, sess.typ
		s.mu.Unlock()

		cm, ok := chatModelOrErr(w, r)
		if !ok {
			s.mu.Lock()
			sess.busy = false
			s.mu.Unlock()
			return
		}
		q, err := graph.OpeningQuestion(r.Context(), cm, profile, typ)

		s.mu.Lock()
		sess.busy = false
		if err != nil {
			s.mu.Unlock()
			writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
			return
		}
		sess.currentQ = q
		sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: q})
		s.mu.Unlock()
		writeOK(w, InterviewMsgResp{Role: "interviewer", Content: q, Turn: 0, Finished: false})
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "content must not be empty")
		return
	}
	if sess.awaiting {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "等待深入考察确认：请先选择「继续深入」或「跳过」")
		return
	}
	if sess.busy {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "正在评估上一题，请稍候")
		return
	}
	if sess.currentQ == "" {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "面试尚未开始（缺少首题）")
		return
	}

	sess.busy = true
	profile, typ := s.profile, sess.typ
	qas := append([]graph.QA(nil), sess.qas...)
	question := sess.currentQ
	turnNo := len(sess.scores) + 1
	s.mu.Unlock()

	cm, ok := chatModelOrErr(w, r)
	if !ok {
		s.mu.Lock()
		sess.busy = false
		s.mu.Unlock()
		return
	}
	result, err := graph.EvaluateTurn(r.Context(), cm, profile, typ, qas, question, req.Content, turnNo, sess.totalTurns, sess.deepDiveUsed)

	s.mu.Lock()
	defer s.mu.Unlock()
	sess.busy = false
	if err != nil {
		writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		return
	}

	sess.transcript = append(sess.transcript, Msg{Role: "user", Content: req.Content})
	sess.qas = append(sess.qas, graph.QA{Question: question, Answer: req.Content})
	note := &AnswerNote{Covered: result.Covered, Missed: result.Missed, Tip: result.Tip}
	sess.reviews = append(sess.reviews, QuestionReview{
		Turn: turnNo - 1, Question: question, Score: result.Score,
		Covered: result.Covered, Missed: result.Missed,
	})
	sess.scores = append(sess.scores, result.Score)
	fb := &Feedback{Score: result.Score, Hint: result.Tip}

	// 模型提议深入考察且本场未用过 → HITL 暂停，缓存常规下一题供「跳过」使用。
	if result.DeepDiveProposal != "" && !sess.deepDiveUsed && !result.Finished && sess.typ == "technical" {
		sess.awaiting = true
		sess.pendingDeepDive = result.DeepDiveProposal
		sess.pendingNext = result.NextQuestion
		writeOK(w, InterviewMsgResp{
			Role: "interviewer", Turn: turnNo - 1, Finished: false,
			AwaitingApproval: true,
			ApprovalPrompt:   fmt.Sprintf("面试官想深入考察「%s」，是否继续？", result.DeepDiveProposal),
			InterimFeedback:  fb, Note: note,
		})
		return
	}

	if result.Finished || len(sess.scores) >= sess.totalTurns {
		writeOK(w, s.finishLocked(sess, note))
		return
	}

	sess.currentQ = result.NextQuestion
	sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: result.NextQuestion})
	writeOK(w, InterviewMsgResp{Role: "interviewer", Content: result.NextQuestion, Turn: turnNo - 1, Finished: false, InterimFeedback: fb, Note: note})
}

func (s *Server) handleInterviewResume(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.findSession(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	if !sess.awaiting {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "当前没有待确认的深入考察")
		return
	}
	if sess.busy {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "正在生成深入考察题，请稍候")
		return
	}
	sess.busy = true
	profile, typ := s.profile, sess.typ
	qas := append([]graph.QA(nil), sess.qas...)
	point := sess.pendingDeepDive
	s.mu.Unlock()

	cm, ok := chatModelOrErr(w, r)
	if !ok {
		s.mu.Lock()
		sess.busy = false
		s.mu.Unlock()
		return
	}
	q, err := graph.DeepDiveQuestion(r.Context(), cm, profile, typ, qas, point)

	s.mu.Lock()
	defer s.mu.Unlock()
	sess.busy = false
	if err != nil {
		writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		return
	}
	sess.awaiting = false
	sess.deepDiveUsed = true
	sess.pendingNext = ""
	sess.pendingDeepDive = ""
	sess.currentQ = q
	sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: q})
	writeOK(w, InterviewMsgResp{Role: "interviewer", Content: q, Turn: len(sess.scores), Finished: false})
}

func (s *Server) handleInterviewSkip(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.findSession(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !sess.awaiting {
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "当前没有待确认的深入考察")
		return
	}
	sess.awaiting = false
	next := sess.pendingNext
	sess.pendingNext = ""
	sess.pendingDeepDive = ""
	sess.currentQ = next
	sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: next})
	writeOK(w, InterviewMsgResp{Role: "interviewer", Content: next, Turn: len(sess.scores), Finished: false})
}

func (s *Server) handleInterviewReport(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.findSession(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	if sess.report != nil {
		s.mu.Unlock()
		writeOK(w, sess.report)
		return
	}
	if !sess.finished {
		s.mu.Unlock()
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "面试尚未结束，暂无报告")
		return
	}
	profile, typ := s.profile, sess.typ
	qas := append([]graph.QA(nil), sess.qas...)
	s.mu.Unlock()

	cm, ok := chatModelOrErr(w, r)
	if !ok {
		return
	}
	result, err := graph.EvaluateInterview(r.Context(), cm, profile, typ, qas)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	sess.report = &InterviewReport{
		SessionID: sess.sessionID, Type: sess.typ,
		OverallScore: result.OverallScore,
		Dimensions:   result.Dimensions,
		Highlights:   result.Highlights,
		Improvements: result.Improvements,
		Transcript:   sess.transcript,
		Review:       sess.reviews,
	}
	writeOK(w, sess.report)
}

// finishLocked 结束面试：写记忆事件与历史记录；调用方须持锁。
func (s *Server) finishLocked(sess *interviewState, note *AnswerNote) InterviewMsgResp {
	sess.finished = true
	sess.transcript = append(sess.transcript, Msg{Role: "interviewer", Content: "好的，本场面试到此结束。"})

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
		overall = overall / len(sess.scores)
	}
	s.intervHist = append(s.intervHist, interviewRecord{SessionID: sess.sessionID, Score: overall, At: nowISO()})
	return InterviewMsgResp{
		Role: "interviewer", Content: "感谢参与！面试已结束，评估报告生成中。", Turn: len(sess.scores), Finished: true,
		InterimFeedback: &Feedback{Score: overall, Hint: "报告已生成，去查看。"}, Note: note,
	}
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
