package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type retryRepository struct {
	job     Job
	created *Job
}

func (r *retryRepository) Get(_ context.Context, _ string) (Job, error) {
	return r.job, nil
}

func (r *retryRepository) Create(_ context.Context, job *Job) error {
	r.created = job

	return nil
}

func (r *retryRepository) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func (r *retryRepository) Select(context.Context, SelectParams) ([]Job, error) {
	return nil, errors.New("not implemented")
}

func (r *retryRepository) Update(context.Context, *Job) error {
	return errors.New("not implemented")
}

func TestRetryCreatesPendingCopy(t *testing.T) {
	repo := &retryRepository{job: Job{
		ID:     "44444444-4444-4444-4444-444444444444",
		Name:   "Da Nang cafes",
		Date:   time.Now().UTC(),
		Status: StatusOK,
		Data: JobData{
			Keywords: []string{"coffee shops in Da Nang"},
			Lang:     "en",
			Zoom:     15,
			Depth:    10,
			MaxTime:  10 * time.Minute,
		},
	}}

	srv, err := New(NewService(repo, t.TempDir()), ":0")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req := requestWithID(httptest.NewRequest(http.MethodPost, "/retry?id="+repo.job.ID, http.NoBody))
	rec := httptest.NewRecorder()
	srv.retry(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if repo.created == nil {
		t.Fatal("expected retry job to be created")
	}

	if repo.created.ID == repo.job.ID || repo.created.Status != StatusPending {
		t.Fatalf("expected a new pending job, got %+v", repo.created)
	}

	if repo.created.Name != "Da Nang cafes (retry)" || len(repo.created.Data.Keywords) != 1 || repo.created.Data.Keywords[0] != repo.job.Data.Keywords[0] {
		t.Fatalf("retry did not preserve source settings: %+v", repo.created)
	}

	if !strings.Contains(rec.Body.String(), `data-job-id="`+repo.created.ID+`"`) {
		t.Fatalf("response did not contain the new job row: %s", rec.Body.String())
	}
}
