package api

import "net/http"

func (s *Server) handleDashboardSummary(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gap := s.computeGap()
	weakDims := make([]string, 0)
	for _, d := range gap.Dimensions {
		if d.Status == "weak" || d.Status == "missing" {
			weakDims = append(weakDims, d.ID)
		}
	}

	radar := s.radarItemsLocked()
	highlights := make([]TitleURL, 0, 3)
	for i := 0; i < len(radar) && i < 3; i++ {
		highlights = append(highlights, TitleURL{Title: radar[i].Title, URL: radar[i].URL})
	}

	writeOK(w, DashboardSummary{
		OverallScore:     gap.OverallScore,
		WeakDimensions:   weakDims,
		RecentExams:      last3Exams(s.examHist),
		RecentInterviews: last3Interviews(s.intervHist),
		StudyStreakDays:  5,
		RadarHighlights:  highlights,
	})
}

func last3Exams(hist []examRecord) []examRecord {
	n := len(hist)
	from := n - 3
	if from < 0 {
		from = 0
	}
	out := make([]examRecord, 0, n-from)
	for i := n - 1; i >= from; i-- {
		out = append(out, hist[i])
	}
	return out
}

func last3Interviews(hist []interviewRecord) []interviewRecord {
	n := len(hist)
	from := n - 3
	if from < 0 {
		from = 0
	}
	out := make([]interviewRecord, 0, n-from)
	for i := n - 1; i >= from; i-- {
		out = append(out, hist[i])
	}
	return out
}
