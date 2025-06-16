package repositories

import (
	"time"
	"twitterservice/internal/domain/entities"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type PostRepository interface {
	Create(post *entities.Post) error
	GetByID(id uuid.UUID) (*entities.Post, error)
	GetByUserID(userID uuid.UUID, limit int) ([]*entities.Post, error)
	Update(post *entities.Post) error
	Delete(id uuid.UUID) error
}

type CassandraPostRepository struct {
	session *gocql.Session
}

func NewCassandraPostRepository(session *gocql.Session) *CassandraPostRepository {
	return &CassandraPostRepository{
		session: session,
	}
}

func (r *CassandraPostRepository) Create(post *entities.Post) error {
	query := `INSERT INTO posts (id, user_id, content, created_at, updated_at, is_deleted) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	return r.session.Query(query,
		post.ID,
		post.UserID,
		post.Content,
		post.CreatedAt,
		post.UpdatedAt,
		post.IsDeleted,
	).Exec()
}

func (r *CassandraPostRepository) GetByID(id uuid.UUID) (*entities.Post, error) {
	var post entities.Post
	query := `SELECT id, user_id, content, created_at, updated_at, is_deleted 
			  FROM posts WHERE id = ? AND is_deleted = false`

	err := r.session.Query(query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.IsDeleted,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

func (r *CassandraPostRepository) GetByUserID(userID uuid.UUID, limit int) ([]*entities.Post, error) {
	var posts []*entities.Post
	query := `SELECT id, user_id, content, created_at, updated_at, is_deleted 
			  FROM posts WHERE user_id = ? AND is_deleted = false 
			  ORDER BY created_at DESC LIMIT ?`

	iter := r.session.Query(query, userID, limit).Iter()
	defer iter.Close()

	for {
		var post entities.Post
		if !iter.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.IsDeleted,
		) {
			break
		}
		posts = append(posts, &post)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *CassandraPostRepository) Update(post *entities.Post) error {
	query := `UPDATE posts SET content = ?, updated_at = ? WHERE id = ?`

	return r.session.Query(query,
		post.Content,
		time.Now(),
		post.ID,
	).Exec()
}

func (r *CassandraPostRepository) Delete(id uuid.UUID) error {
	query := `UPDATE posts SET is_deleted = true, updated_at = ? WHERE id = ?`

	return r.session.Query(query, time.Now(), id).Exec()
}
