package web

import (
	"net/http"

	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
)

type CommentHandler struct {
	store goreddit.Store
}

func (h *CommentHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := getId(r, "postID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		content := r.FormValue("content")

		if err := h.store.CreateComment(&goreddit.Comment{
			ID:      uuid.New(),
			PostID:  postID,
			Content: content,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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

		if err := h.store.UpdateComment(&goreddit.Comment{
			ID:      c.ID,
			PostID:  c.PostID,
			Content: c.Content,
			Votes:   c.Votes + newVoteAmount,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
