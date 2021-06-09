package postgres

import (
	"fmt"

	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CommentStore struct {
	*sqlx.DB
}

func (s *CommentStore) Comment(id uuid.UUID) (goreddit.Comment, error) {
	var c goreddit.Comment
	if err := s.Get(&c, `SELECT * FROM comments WHERE id = $1`, id); err != nil {
		return goreddit.Comment{}, fmt.Errorf("error getting comment: %w", err)
	}

	return c, nil
}

func (s *CommentStore) CommentsByPost(postID uuid.UUID) ([]goreddit.Comment, error) {
	var ps []goreddit.Comment
	if err := s.Select(&ps, `SELECT * FROM comments WHERE thread_id = $1`, postID); err != nil {
		return []goreddit.Comment{}, fmt.Errorf("error getting comments: %w", err)
	}

	return ps, nil
}

func (s *CommentStore) CreateComment(c *goreddit.Comment) error {
	if err := s.Get(c, `INSERT INTO comments VALUES ($1, $2, $3, $4) RETURNING *`, c.ID, c.PostID, c.Content, c.Votes); err != nil {
		return fmt.Errorf("error creating comment: %w", err)
	}

	return nil
}

func (s *CommentStore) UpdateComment(c *goreddit.Comment) error {
	if err := s.Get(c, `UPDATE comments SET post_id = $1, content = $2, votes = $3 WHERE id = $4 RETURNING *`, c.PostID, c.Content, c.Votes, c.ID); err != nil {
		return fmt.Errorf("error updating comment: %w", err)
	}

	return nil
}

func (s *CommentStore) DeleteComment(id uuid.UUID) error {
	if _, err := s.Exec(`DELETE FROM comments WHERE id = $1`, id); err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}

	return nil
}
