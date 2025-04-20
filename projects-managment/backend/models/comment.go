package models

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type Comment struct {
	ID         int       `json:"id"`
	ProjectID  int       `json:"projectId"`
	UserID     int       `json:"userId"`
	Content    string    `json:"content"`
	User       *User     `json:"user,omitempty"`
	LikesCount int       `json:"likesCount"`
	IsLiked    bool      `json:"isLiked"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type CommentService struct {
	DB *sql.DB
}

type CommentError struct {
	Code    int
	Message string
}

func (e *CommentError) Error() string {
	return e.Message
}

var (
	ErrCommentNotFound = &CommentError{Code: http.StatusNotFound, Message: "comment not found"}
)

func (cs *CommentService) Create(comment Comment) (Comment, error) {
	var commentID int
	err := cs.DB.QueryRow(
		"INSERT INTO comments (project_id, user_id, content) VALUES ($1, $2, $3) RETURNING id",
		comment.ProjectID, comment.UserID, comment.Content,
	).Scan(&commentID)

	if err != nil {
		return comment, fmt.Errorf("failed to create comment: %v", err)
	}

	comment.ID = commentID
	return comment, nil
}

func (cs *CommentService) GetProjectComments(projectID int, currentUserID int) ([]Comment, error) {
	query := `
		SELECT c.id, c.project_id, c.user_id, c.content, u.username, c.likes_count, 
		       CASE WHEN cl.user_id IS NOT NULL THEN true ELSE false END as is_liked,
		       c.created_at, c.updated_at 
		FROM comments c 
		JOIN users u ON c.user_id = u.id
		LEFT JOIN comment_likes cl ON c.id = cl.comment_id AND cl.user_id = $2
		WHERE c.project_id = $1
		ORDER BY c.created_at DESC`

	rows, err := cs.DB.Query(query, projectID, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		var comment Comment
		var username string
		comment.User = &User{}

		err := rows.Scan(
			&comment.ID, &comment.ProjectID, &comment.UserID, &comment.Content,
			&username, &comment.LikesCount, &comment.IsLiked, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to parse comment data: %v", err)
		}

		comment.User.Username = username
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %v", err)
	}

	return comments, nil
}

func (cs *CommentService) LikeComment(commentID, userID int) error {
	tx, err := cs.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE comment_id = $1 AND user_id = $2)",
		commentID, userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}
	if exists {
		return &CommentError{Code: http.StatusBadRequest, Message: "comment already liked by user"}
	}

	// Insert like
	_, err = tx.Exec("INSERT INTO comment_likes (comment_id, user_id) VALUES ($1, $2)",
		commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to like comment: %v", err)
	}

	// Update likes count
	_, err = tx.Exec("UPDATE comments SET likes_count = likes_count + 1 WHERE id = $1", commentID)
	if err != nil {
		return fmt.Errorf("failed to update likes count: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (cs *CommentService) UnlikeComment(commentID, userID int) error {
	tx, err := cs.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var exists bool

	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM comment_likes WHERE comment_id = $1 AND user_id = $2)",
		commentID, userID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("database error: %v", err)
	}
	if !exists {
		return &CommentError{Code: http.StatusBadRequest, Message: "comment not liked by user"}
	}

	// Delete like
	result, err := tx.Exec("DELETE FROM comment_likes WHERE comment_id = $1 AND user_id = $2",
		commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to unlike comment: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return &CommentError{Code: http.StatusBadRequest, Message: "comment not liked by user"}
	}

	// Update likes count
	_, err = tx.Exec("UPDATE comments SET likes_count = likes_count - 1 WHERE id = $1", commentID)
	if err != nil {
		return fmt.Errorf("failed to update likes count: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (cs *CommentService) Delete(commentID, userID int) error {
	_, err := cs.DB.Exec("DELETE FROM comments WHERE id = $1 AND user_id = $2", commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %v", err)
	}

	return nil
}
