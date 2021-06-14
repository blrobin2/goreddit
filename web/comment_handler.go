package web

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
)

type CommentHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *CommentHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getId(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		form := CreateCommentForm{
			Content: r.FormValue("content"),
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(w, r, r.Referer(), http.StatusFound)
			return
		}

		if err := h.store.CreateComment(&goreddit.Comment{
			ID:      uuid.New(),
			PostID:  postID,
			Content: form.Content,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "Your comment has been submitted.")

		http.Redirect(w, r, "/posts/"+postID.String(), http.StatusFound)
	}
}

func (h *CommentHandler) Upvote() http.HandlerFunc {
	return voteOnComment(h, 1)
}

func (h *CommentHandler) Downvote() http.HandlerFunc {
	return voteOnComment(h, -1)
}

func voteOnComment(h *CommentHandler, newVoteAmount int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getId(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		c, err := h.store.Comment(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c.Votes += newVoteAmount
		if err := h.store.UpdateComment(&c); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
