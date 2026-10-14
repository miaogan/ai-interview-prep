package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"aiprep/backend/internal/graph"
	llmmodel "aiprep/backend/internal/model"
	"aiprep/backend/internal/resume"
)

func (s *Server) handleUploadResume(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "invalid multipart form: "+err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "missing form field 'file'")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 20<<20))
	if err != nil {
		writeErr(w, http.StatusBadRequest, CodeBadRequest, "读取文件失败: "+err.Error())
		return
	}

	text, err := resume.ExtractText(header.Filename, data)
	if err != nil {
		writeErr(w, http.StatusBadRequest, CodeParseFailed, err.Error())
		return
	}

	cm, err := llmmodel.NewChatModel(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		return
	}
	p, err := graph.ParseResume(r.Context(), cm, text)
	if err != nil {
		if errors.Is(err, graph.ErrInvalidResume) {
			writeErr(w, http.StatusBadRequest, CodeParseFailed, err.Error())
		} else {
			writeErr(w, http.StatusInternalServerError, CodeModelFailed, err.Error())
		}
		return
	}

	uploadedAt := nowISO()
	p.UpdatedAt = uploadedAt

	s.mu.Lock()
	s.profile = p
	s.mu.Unlock()

	writeOK(w, ResumeUploadResp{
		ResumeID:      fmt.Sprintf("r_%d", time.Now().UnixMilli()),
		FileName:      header.Filename,
		ParsedProfile: p,
		UploadedAt:    uploadedAt,
	})
}

func (s *Server) handleGetProfile(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeOK(w, s.profile) // 未上传时 data 为 null，与前端 mock 一致
}
