package web

import (
	"net/http"
	"text/template"

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

	h.Route("/threads", func(r chi.Router) {
		r.Get("/", h.ThreadsList())
		r.Get("/new", h.NewThread())
		r.Post("/", h.CreateThread())
		r.Delete("/{id}", h.DeleteThread())
	})

	return h
}

type Handler struct {
	*chi.Mux

	store goreddit.Store
}

const threadListHTML = `
<h1>Threads</h1>
<ul>
{{range .Threads}}
	<li>
		<h3>{{.Title}}</h3>
		<p>{{.Description}}</p>
		<p>
			<button onclick="handleDelete('{{.ID}}')">Delete</button>
		</p>
	</li>
{{end}}
</ul>
<a href="/threads/new">Create New Threadz</a>
<script>
	function handleDelete(id) {
		return fetch("/threads/" + id, {
			method: 'DELETE',
		}).then(res => {
			window.location.reload();
		});
	}
</script>
`

func (h *Handler) ThreadsList() http.HandlerFunc {
	type data struct {
		Threads []goreddit.Thread
	}
	templ := template.Must(template.New("").Parse(threadListHTML))
	return func(w http.ResponseWriter, r *http.Request) {
		ts, err := h.store.Threads()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templ.Execute(w, data{Threads: ts})
	}
}

const newThreadHTML = `
<h1>New thread</h1>
<form action="/threads" method="POST">
	<p>
		<label for="title">Title</label>
		<input type="text" name="title" id="title" />
	</p>

	<p>
		<label for="description">Description</label>
		<textarea name="description" id="description"></textarea>
	</p>

	<p>
		<button type="submit">
			Create Thread
		</button>
	</p>
</form>
`

func (h *Handler) NewThread() http.HandlerFunc {
	templ := template.Must(template.New("").Parse(newThreadHTML))
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Execute(w, nil)
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
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
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
