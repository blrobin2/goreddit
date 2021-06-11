package web

import (
	"html/template"
	"net/http"

	"github.com/blrobin2/goreddit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func NewHandler(store goreddit.Store) *Handler {
	h := &Handler{
		Mux:   chi.NewMux(),
		store: store,
	}

	h.Use(middleware.Logger)

	h.Get("/", h.Home())
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", h.ThreadsList())
		r.Get("/new", h.NewThread())
		r.Post("/", h.CreateThread())
		r.Get("/{id}", h.ShowThread())
		r.Delete("/{id}", h.DeleteThread())

		r.Get("/{id}/new", h.NewPost())
		r.Post("/{id}", h.CreatePost())
		r.Get("/{threadID}/{postID}", h.ShowPost())
	})

	return h
}

type Handler struct {
	*chi.Mux

	store goreddit.Store
}

func (h *Handler) Home() http.HandlerFunc {
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Execute(w, nil)
	}
}

func (h *Handler) ThreadsList() http.HandlerFunc {
	type data struct {
		Threads []goreddit.Thread
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/threads.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		ts, err := h.store.Threads()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Execute(w, data{Threads: ts})
	}
}

func (h *Handler) NewThread() http.HandlerFunc {
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/thread_create.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Execute(w, nil)
	}
}

func (h *Handler) ShowThread() http.HandlerFunc {
	type data struct {
		Thread goreddit.Thread
		Posts  []goreddit.Post
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/thread.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getId(r, "id")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		t, err := h.store.Thread(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ps, err := h.store.PostsByThead(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Execute(w, data{
			Thread: t,
			Posts:  ps,
		})
	}
}

func (h *Handler) CreateThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.FormValue("title")
		description := r.FormValue("description")

		if err := h.store.CreateThread(&goreddit.Thread{
			ID:          uuid.New(),
			Title:       title,
			Description: description,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/threads", http.StatusFound)
	}
}

func (h *Handler) DeleteThread() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getId(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := h.store.DeleteThread(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) NewPost() http.HandlerFunc {
	type data struct {
		Thread goreddit.Thread
	}

	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/post_create.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getId(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t, err := h.store.Thread(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Execute(w, data{
			Thread: t,
		})
	}
}

func (h *Handler) ShowPost() http.HandlerFunc {
	type data struct {
		Thread   goreddit.Thread
		Post     goreddit.Post
		Comments []goreddit.Comment
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/post.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		threadID, err := getId(r, "threadID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		postID, err := getId(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t, err := h.store.Thread(threadID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p, err := h.store.Post(postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cs, err := h.store.CommentsByPost(postID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Execute(w, data{
			Thread:   t,
			Post:     p,
			Comments: cs,
		})
	}
}

func (h *Handler) CreatePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getId(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		content := r.FormValue("content")

		p := &goreddit.Post{
			ID:       uuid.New(),
			ThreadID: id,
			Title:    title,
			Content:  content,
		}
		if err := h.store.CreatePost(p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/threads/"+id.String()+"/"+p.ID.String(), http.StatusFound)
	}
}

func getId(r *http.Request, idName string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, idName)
	return uuid.Parse(idStr)
}
