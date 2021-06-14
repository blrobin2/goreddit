package web

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/alexedwards/scs/v2"
	"github.com/blrobin2/goreddit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

func NewHandler(store goreddit.Store, sessions *scs.SessionManager, csrfKey []byte) *Handler {
	h := &Handler{
		Mux:      chi.NewMux(),
		store:    store,
		sessions: sessions,
	}

	threads := ThreadHandler{store: store, sessions: sessions}
	posts := PostHandler{store: store, sessions: sessions}
	comments := CommentHandler{store: store, sessions: sessions}

	h.Use(middleware.Logger)
	h.Use(csrf.Protect(csrfKey, csrf.Secure(false)))
	h.Use(sessions.LoadAndSave)

	h.Get("/", h.Home())
	h.Route("/threads", func(r chi.Router) {
		r.Get("/", threads.List())
		r.Get("/new", threads.New())
		r.Post("/", threads.Create())
		r.Get("/{id}", threads.Show())
		r.Delete("/{id}", threads.Delete())

		r.Get("/{id}/new", posts.New())
		r.Post("/{id}", posts.Create())
	})
	h.Route("/posts", func(r chi.Router) {
		r.Get("/{postID}", posts.Show())
		r.Post("/{postID}", comments.Create())
		r.Post("/{id}/upvote", posts.Upvote())
		r.Post("/{id}/downvote", posts.Downvote())
	})

	h.Route("/comments", func(r chi.Router) {
		r.Post("/{id}/upvote", comments.Upvote())
		r.Post("/{id}/downvote", comments.Downvote())
	})

	return h
}

type Handler struct {
	*chi.Mux

	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *Handler) Home() http.HandlerFunc {
	type data struct {
		SessionData
		CSRFToken string
		Posts     []goreddit.Post
	}

	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/home.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		ps, err := h.store.Posts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sort.SliceStable(ps, func(i, j int) bool {
			return ps[i].Votes > ps[j].Votes
		})

		templ.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			CSRFToken:   csrf.Token(r),
			Posts:       ps,
		})
	}
}

func getId(r *http.Request, idName string) (uuid.UUID, error) {
	idStr := chi.URLParam(r, idName)
	return uuid.Parse(idStr)
}
