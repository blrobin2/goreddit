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
		SessionData SessionData
		Threads     []goreddit.Thread
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/threads.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		ts, err := h.store.Threads()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			Threads:     ts,
		})
	}
}

func (h *ThreadHandler) New() http.HandlerFunc {
	type data struct {
		SessionData
		CSRF template.HTML
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/thread_create.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			CSRF:        csrf.TemplateField(r),
		})
	}
}

func (h *ThreadHandler) Show() http.HandlerFunc {
	type data struct {
		SessionData
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
			SessionData: GetSessionData(h.sessions, r.Context()),
			CSRFToken:   csrf.Token(r),
			Thread:      t,
			Posts:       ps,
		})
	}
}

func (h *ThreadHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form := CreateThreadForm{
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(w, r, r.Referer(), http.StatusFound)
			return
		}

		if err := h.store.CreateThread(&goreddit.Thread{
			ID:          uuid.New(),
			Title:       form.Title,
			Description: form.Description,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "Your thread has been created.")

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

		h.sessions.Put(r.Context(), "flash", "Your thread has been deleted.")

		w.WriteHeader(http.StatusNoContent)
	}
}
