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

	// Convert google/uuid to gocql.UUID
	postID, _ := gocql.ParseUUID(post.ID.String())
	userID, _ := gocql.ParseUUID(post.UserID.String())

	return r.session.Query(query,
		postID,
		userID,
		post.Content,
		post.CreatedAt,
		post.UpdatedAt,
		post.IsDeleted,
	).Exec()
}

func (r *CassandraPostRepository) GetByID(id uuid.UUID) (*entities.Post, error) {
	var post entities.Post
	var postID, userID gocql.UUID
	query := `SELECT id, user_id, content, created_at, updated_at, is_deleted 
			  FROM posts WHERE id = ? AND is_deleted = false`

	// Convert google/uuid to gocql.UUID for query
	queryID, _ := gocql.ParseUUID(id.String())

	err := r.session.Query(query, queryID).Scan(
		&postID,
		&userID,
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

	// Convert gocql.UUID back to google/uuid
	post.ID, _ = uuid.Parse(postID.String())
	post.UserID, _ = uuid.Parse(userID.String())

	return &post, nil
}

func (r *CassandraPostRepository) GetByUserID(userID uuid.UUID, limit int) ([]*entities.Post, error) {
	var posts []*entities.Post
	query := `SELECT id, user_id, content, created_at, updated_at, is_deleted 
			  FROM posts WHERE user_id = ? AND is_deleted = false 
			  ORDER BY created_at DESC LIMIT ?`

	// Convert google/uuid to gocql.UUID for query
	queryUserID, _ := gocql.ParseUUID(userID.String())

	iter := r.session.Query(query, queryUserID, limit).Iter()
	defer iter.Close()

	for {
		var post entities.Post
		var postID, postUserID gocql.UUID
		if !iter.Scan(
			&postID,
			&postUserID,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.IsDeleted,
		) {
			break
		}

		// Convert gocql.UUID back to google/uuid
		post.ID, _ = uuid.Parse(postID.String())
		post.UserID, _ = uuid.Parse(postUserID.String())

		posts = append(posts, &post)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *CassandraPostRepository) Update(post *entities.Post) error {
	query := `UPDATE posts SET content = ?, updated_at = ? WHERE id = ? AND is_deleted = ?`

	// Convert google/uuid to gocql.UUID
	postID, _ := gocql.ParseUUID(post.ID.String())

	return r.session.Query(query,
		post.Content,
		time.Now(),
		postID,
		post.IsDeleted,
	).Exec()
}

func (r *CassandraPostRepository) Delete(id uuid.UUID) error {
	// First, get the existing post to copy its data
	existingPost, err := r.GetByID(id)
	if err != nil {
		return err
	}
	if existingPost == nil {
		return nil // Post not found, consider it already deleted
	}

	// Convert google/uuid to gocql.UUID
	postID, _ := gocql.ParseUUID(id.String())
	userID, _ := gocql.ParseUUID(existingPost.UserID.String())

	// Delete the active record (id, is_deleted = false)
	deleteQuery := `DELETE FROM posts WHERE id = ? AND is_deleted = false`
	if err := r.session.Query(deleteQuery, postID).Exec(); err != nil {
		return err
	}

	// Insert the deleted record (id, is_deleted = true)
	insertQuery := `INSERT INTO posts (id, user_id, content, created_at, updated_at, is_deleted) 
					VALUES (?, ?, ?, ?, ?, ?)`

	return r.session.Query(insertQuery,
		postID,
		userID,
		existingPost.Content,
		existingPost.CreatedAt,
		time.Now(),
		true,
	).Exec()
}
