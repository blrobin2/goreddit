package web

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/alexedwards/scs/v2"
	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
)

type ThreadHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *ThreadHandler) List() http.HandlerFunc {
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

func (h *ThreadHandler) New() http.HandlerFunc {
	type data struct {
		CSRF template.HTML
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/thread_create.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Execute(w, data{
			CSRF: csrf.TemplateField(r),
		})
	}
}

func (h *ThreadHandler) Show() http.HandlerFunc {
	type data struct {
		CSRFToken string
		Thread    goreddit.Thread
		Posts     []goreddit.Post
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

		sort.SliceStable(ps, func(i, j int) bool {
			return ps[i].Votes > ps[j].Votes
		})

		templ.Execute(w, data{
			CSRFToken: csrf.Token(r),
			Thread:    t,
			Posts:     ps,
		})
	}
}

func (h *ThreadHandler) Create() http.HandlerFunc {
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

func (h *ThreadHandler) Delete() http.HandlerFunc {
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
