package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PostRequest struct {
	Title string `json:"title"`
}

type PatchRequest struct {
	Done *bool `json:"done"`
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{clock: clock, seq: 0, tasks: make(map[string]Task)}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return Task{}, ErrInvalidTitle
	}

	task := Task{
		ID:        strconv.FormatUint(r.seq, 10),
		Title:     trimmedTitle,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}
	r.tasks[task.ID] = task
	r.seq++
	return task, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	return task, ok
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	filtered := make([]Task, 0)

	for _, task := range r.tasks {
		if task.Done == done {
			filtered = append(filtered, task)
		}
	}

	return filtered
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if task.Done != done {
		task.Done = done
		task.UpdatedAt = r.clock.Now()
		r.tasks[task.ID] = task
	}

	return task, nil
}

type httpHandler struct {
	repo TaskRepo
	mux  *http.ServeMux
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	h := &httpHandler{repo: repo, mux: http.NewServeMux()}

	h.mux.HandleFunc("POST /tasks", h.handleCreate)
	h.mux.HandleFunc("GET /tasks", h.handleList)
	h.mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		h.handleGet(w, r, r.PathValue("id"))
	})
	h.mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		h.handlePatch(w, r, r.PathValue("id"))
	})
	return h
}
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	postRequest := PostRequest{}
	if err := decodeStrictJSON(r.Body, &postRequest); err != nil {
		writeJSON(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.repo.Create(postRequest.Title)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTitle):
			writeJSON(w, http.StatusBadRequest, err)
		default:
			writeJSON(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, task)
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	if task, ok := h.repo.Get(id); ok {
		writeJSON(w, http.StatusOK, task)
	} else {
		writeJSON(w, http.StatusNotFound, nil)
	}
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	doneValues := r.URL.Query()["done"]
	done := false

	switch len(doneValues) {
	case 0:
	case 1:
		switch doneValues[0] {
		case "true":
			done = true
		case "false":
			done = false
		default:
			writeJSON(w, http.StatusBadRequest, "done must be true or false")
			return
		}
	default:
		writeJSON(w, http.StatusBadRequest, "done must be specified once")
		return
	}

	tasks := h.repo.List(done)
	slices.SortFunc(tasks, func(a, b Task) int {
		if c := b.UpdatedAt.Compare(a.UpdatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})

	writeJSON(w, http.StatusOK, tasks)
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	patchRequest := PatchRequest{}
	if err := decodeStrictJSON(r.Body, &patchRequest); err != nil {
		writeJSON(w, http.StatusBadRequest, err)
		return
	}

	if patchRequest.Done == nil {
		writeJSON(w, http.StatusBadRequest, nil)
		return
	}

	task, err := h.repo.SetDone(id, *patchRequest.Done)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeJSON(w, http.StatusNotFound, nil)
		default:
			writeJSON(w, http.StatusInternalServerError, err)
		}
		return
	}
	writeJSON(w, http.StatusOK, task)
}
func decodeStrictJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}

	if dec.More() {
		return errors.New("unexpected trailing data")
	}

	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
