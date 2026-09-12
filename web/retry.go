package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *Server) retry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	id, ok := getIDFromRequest(r)
	if !ok {
		http.Error(w, "Invalid ID", http.StatusUnprocessableEntity)

		return
	}

	original, err := s.svc.Get(r.Context(), id.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("Job not found: %v", err), http.StatusNotFound)

		return
	}

	if original.Status != StatusOK && original.Status != StatusFailed {
		http.Error(w, "Only completed or failed jobs can be retried", http.StatusConflict)

		return
	}

	newJob := Job{
		ID:     uuid.New().String(),
		Name:   original.Name + " (retry)",
		Date:   time.Now().UTC(),
		Status: StatusPending,
		Data: JobData{
			Keywords:     append([]string(nil), original.Data.Keywords...),
			Lang:         original.Data.Lang,
			Zoom:         original.Data.Zoom,
			Lat:          original.Data.Lat,
			Lon:          original.Data.Lon,
			FastMode:     original.Data.FastMode,
			Radius:       original.Data.Radius,
			Depth:        original.Data.Depth,
			Email:        original.Data.Email,
			ExtraReviews: original.Data.ExtraReviews,
			MaxTime:      original.Data.MaxTime,
			Proxies:      append([]string(nil), original.Data.Proxies...),
		},
	}

	if err := newJob.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Invalid source job: %v", err), http.StatusUnprocessableEntity)

		return
	}

	if err := s.svc.Create(r.Context(), &newJob); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	tmpl, ok := s.tmpl["static/templates/job_row.html"]
	if !ok {
		http.Error(w, "missing tpl", http.StatusInternalServerError)

		return
	}

	_ = tmpl.Execute(w, newJob)
}
