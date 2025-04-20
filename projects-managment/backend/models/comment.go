package models

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type Comment struct {
	ID        int       `json:"id"`
	ProjectID int       `json:"projectId"`
	UserID    int       `json:"userId"`
	Content   string    `json:"content"`
	User      *User     `json:"user,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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

func (cs *CommentService) GetProjectComments(projectID int) ([]Comment, error) {
	var projectCount int
	err := cs.DB.QueryRow("SELECT COUNT(*) FROM projects WHERE id = $1", projectID).Scan(&projectCount)
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	if projectCount == 0 {
		return nil, &CommentError{Code: http.StatusNotFound, Message: "project not found"}
	}

	rows, err := cs.DB.Query(
		`SELECT c.id, c.project_id, c.user_id, c.content, u.username, c.created_at, c.updated_at 
		 FROM comments c 
		 JOIN users u ON c.user_id = u.id
		 WHERE c.project_id = $1
		 ORDER BY c.created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var username string
		comment.User = &User{}

		err := rows.Scan(
			&comment.ID, &comment.ProjectID, &comment.UserID, &comment.Content,
			&username, &comment.CreatedAt, &comment.UpdatedAt,
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

func (cs *CommentService) Delete(commentID, userID int) error {
	_, err := cs.DB.Exec("DELETE FROM comments WHERE id = $1 AND user_id = $2", commentID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %v", err)
	}

	return nil
}
