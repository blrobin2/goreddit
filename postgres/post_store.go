package postgres

import (
	"fmt"

	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PostStore struct {
	*sqlx.DB
}

func (s *PostStore) Post(id uuid.UUID) (goreddit.Post, error) {
	var p goreddit.Post
	if err := s.Get(&p, `SELECT * FROM posts WHERE id = $1`, id); err != nil {
		return goreddit.Post{}, fmt.Errorf("error getting post: %w", err)
	}

	return p, nil
}

func (s *PostStore) PostsByThead(threadID uuid.UUID) ([]goreddit.Post, error) {
	var ps []goreddit.Post
	query := `
	SELECT
		posts.*,
		COUNT(comments.*) AS comments_count
	FROM posts
	LEFT JOIN comments ON comments.post_id = posts.id
	WHERE thread_id = $1
	GROUP BY posts.id
	`
	if err := s.Select(&ps, query, threadID); err != nil {
		return []goreddit.Post{}, fmt.Errorf("error getting posts: %w", err)
	}

	return ps, nil
}

func (s *PostStore) Posts() ([]goreddit.Post, error) {
	var ps []goreddit.Post
	query := `
	SELECT
		posts.*,
		threads.title AS thread_title,
		COUNT(comments.*) AS comments_count
	FROM posts
	JOIN threads ON posts.thread_id = threads.id
	LEFT JOIN comments ON comments.post_id = posts.id
	GROUP BY posts.id, threads.title
	`
	if err := s.Select(&ps, query); err != nil {
		return []goreddit.Post{}, fmt.Errorf("error getting posts: %w", err)
	}

	return ps, nil
}

func (s *PostStore) CreatePost(p *goreddit.Post) error {
	if err := s.Get(p, `INSERT INTO posts VALUES ($1, $2, $3, $4, $5) RETURNING *`, p.ID, p.ThreadID, p.Title, p.Content, p.Votes); err != nil {
		return fmt.Errorf("error creating post: %w", err)
	}

	return nil
}

func (s *PostStore) UpdatePost(p *goreddit.Post) error {
	if err := s.Get(p, `UPDATE posts SET thread_id = $1, title = $2, content = $3, votes = $4 WHERE id = $5 RETURNING *`, p.ThreadID, p.Title, p.Content, p.Votes, p.ID); err != nil {
		return fmt.Errorf("error updating post: %w", err)
	}

	return nil
}

func (s *PostStore) DeletePost(id uuid.UUID) error {
	if _, err := s.Exec(`DELETE FROM posts WHERE id = $1`, id); err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}

	return nil
}
