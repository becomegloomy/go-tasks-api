package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/becomegloomy/go-tasks-api/internal/handler"
	"github.com/becomegloomy/go-tasks-api/internal/model"
	"github.com/becomegloomy/go-tasks-api/internal/storage"
)

// memStore — in-memory реализация storage.Store для тестов.
type memStore struct {
	tasks  map[int64]model.Task
	nextID int64
}

func newMemStore() *memStore {
	return &memStore{tasks: map[int64]model.Task{}, nextID: 1}
}

func (m *memStore) List(ctx context.Context) ([]model.Task, error) {
	var out []model.Task
	for _, t := range m.tasks {
		out = append(out, t)
	}
	return out, nil
}

func (m *memStore) Get(ctx context.Context, id int64) (model.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return model.Task{}, storage.ErrNotFound
	}
	return t, nil
}

func (m *memStore) Create(ctx context.Context, title string) (model.Task, error) {
	t := model.Task{ID: m.nextID, Title: title}
	m.tasks[t.ID] = t
	m.nextID++
	return t, nil
}

func (m *memStore) Update(ctx context.Context, id int64, title string, done bool) (model.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return model.Task{}, storage.ErrNotFound
	}
	t.Title = title
	t.Done = done
	m.tasks[id] = t
	return t, nil
}

func (m *memStore) Delete(ctx context.Context, id int64) error {
	if _, ok := m.tasks[id]; !ok {
		return storage.ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func (m *memStore) Close() error { return nil }

func newServer() http.Handler {
	h := handler.New(newMemStore())
	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", h.List)
	mux.HandleFunc("POST /tasks", h.Create)
	mux.HandleFunc("GET /tasks/{id}", h.Get)
	mux.HandleFunc("PUT /tasks/{id}", h.Update)
	mux.HandleFunc("DELETE /tasks/{id}", h.Delete)
	return mux
}

func TestCreateAndGet(t *testing.T) {
	srv := newServer()

	body := bytes.NewBufferString(`{"title":"write tests"}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d", rec.Code)
	}

	var created model.Task
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID == 0 || created.Title != "write tests" {
		t.Fatalf("unexpected task: %+v", created)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: want 200, got %d", rec.Code)
	}
}

func TestCreateInvalidBody(t *testing.T) {
	srv := newServer()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{}`))
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	srv := newServer()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestDelete(t *testing.T) {
	srv := newServer()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"title":"x"}`))
	srv.ServeHTTP(rec, req)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("after delete: want 404, got %d", rec.Code)
	}
}
