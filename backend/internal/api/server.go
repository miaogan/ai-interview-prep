package api

import (
	"net/http"
	"sync"
	"time"
)

func nowISO() string { return time.Now().UTC().Format(time.RFC3339) }

// Server 持有全部进程内状态；骨架阶段用内存替代 SQLite（见 docs/03 记忆层）。
type Server struct {
	mu sync.Mutex

	profile    *Profile
	memEvents  []memoryEvent
	exam       *examSession
	examHist   []examRecord
	sessions   map[string]*interviewState
	intervHist []interviewRecord
}

type memoryEvent struct {
	Type        string `json:"type"` // exam | interview
	DimensionID string `json:"dimensionId"`
	Delta       int    `json:"delta"`
}

type examRecord struct {
	ExamID string `json:"examId"`
	Score  int    `json:"score"`
	At     string `json:"at"`
}

type interviewRecord struct {
	SessionID string `json:"sessionId"`
	Score     int    `json:"score"`
	At        string `json:"at"`
}

func NewRouter() http.Handler {
	s := &Server{sessions: map[string]*interviewState{}}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/resume/upload", s.handleUploadResume)
	mux.HandleFunc("GET /api/v1/resume/profile", s.handleGetProfile)
	mux.HandleFunc("GET /api/v1/gap/analysis", s.handleGapAnalysis)
	mux.HandleFunc("POST /api/v1/exam/generate", s.handleGenerateExam)
	mux.HandleFunc("POST /api/v1/exam/{examId}/submit", s.handleSubmitExam)
	mux.HandleFunc("POST /api/v1/interview/session", s.handleCreateInterview)
	mux.HandleFunc("POST /api/v1/interview/{sessionId}/message", s.handleInterviewMessage)
	mux.HandleFunc("POST /api/v1/interview/{sessionId}/resume", s.handleInterviewResume)
	mux.HandleFunc("POST /api/v1/interview/{sessionId}/skip", s.handleInterviewSkip)
	mux.HandleFunc("GET /api/v1/interview/{sessionId}/report", s.handleInterviewReport)
	mux.HandleFunc("GET /api/v1/radar/latest", s.handleRadarLatest)
	mux.HandleFunc("GET /api/v1/dashboard/summary", s.handleDashboardSummary)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, http.StatusNotFound, CodeNotFound, "resource not found: "+r.URL.Path)
	})

	return cors(mux)
}

// cors 允许前端 vite dev(5173) 跨域调用；MVP 单用户，放开 Origin。
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
