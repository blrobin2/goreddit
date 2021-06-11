package web

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
)

type PostHandler struct {
	store goreddit.Store
}

func (h *PostHandler) New() http.HandlerFunc {
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

func (h *PostHandler) Create() http.HandlerFunc {
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

		http.Redirect(w, r, "/posts/"+p.ID.String(), http.StatusFound)
	}
}

func (h *PostHandler) Show() http.HandlerFunc {
	type data struct {
		Post     goreddit.Post
		Comments []goreddit.Comment
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/post.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getId(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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

		sort.SliceStable(cs, func(i, j int) bool {
			return cs[i].Votes > cs[j].Votes
		})

		templ.Execute(w, data{
			Post:     p,
			Comments: cs,
		})
	}
}

func (h *PostHandler) Upvote() http.HandlerFunc {
	return voteOnPost(h, 1)
}

func (h *PostHandler) Downvote() http.HandlerFunc {
	return voteOnPost(h, -1)
}

func voteOnPost(h *PostHandler, newVoteAmount int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getId(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		p, err := h.store.Post(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p.Votes += newVoteAmount
		if err := h.store.UpdatePost(&p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
